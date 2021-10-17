package main

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/Luzifer/twitch-bot/twitch"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	rawMessageHandlers     []plugins.RawMessageHandlerFunc
	rawMessageHandlersLock sync.Mutex
)

func notifyRawMessageHandlers(m *irc.Message) error {
	rawMessageHandlersLock.Lock()
	defer rawMessageHandlersLock.Unlock()

	for _, fn := range rawMessageHandlers {
		if err := fn(m); err != nil {
			return errors.Wrap(err, "executing raw message handlers")
		}
	}

	return nil
}

func registerRawMessageHandler(fn plugins.RawMessageHandlerFunc) error {
	rawMessageHandlersLock.Lock()
	defer rawMessageHandlersLock.Unlock()

	rawMessageHandlers = append(rawMessageHandlers, fn)

	return nil
}

type ircHandler struct {
	conn *tls.Conn
	c    *irc.Client
	user string
}

func newIRCHandler() (*ircHandler, error) {
	h := new(ircHandler)

	username, err := twitchClient.GetAuthorizedUsername()
	if err != nil {
		return nil, errors.Wrap(err, "fetching username")
	}

	conn, err := tls.Dial("tcp", "irc.chat.twitch.tv:6697", nil)
	if err != nil {
		return nil, errors.Wrap(err, "connect to IRC server")
	}

	h.c = irc.NewClient(conn, irc.ClientConfig{
		Nick:    username,
		Pass:    strings.Join([]string{"oauth", cfg.TwitchToken}, ":"),
		User:    username,
		Name:    username,
		Handler: h,

		SendLimit: cfg.IRCRateLimit,
		SendBurst: 0, // Twitch uses a bucket system, we don't have anything to replicate that in this IRC client
	})
	h.conn = conn
	h.user = username

	return h, nil
}

func (i ircHandler) Client() *irc.Client { return i.c }

func (i ircHandler) Close() error { return i.conn.Close() }

func (i ircHandler) ExecuteJoins(channels []string) {
	for _, ch := range channels {
		i.c.Write(fmt.Sprintf("JOIN #%s", strings.TrimLeft(ch, "#")))
	}
}

func (i ircHandler) ExecutePart(channel string) {
	i.c.Write(fmt.Sprintf("PART #%s", strings.TrimLeft(channel, "#")))
}

func (i ircHandler) Handle(c *irc.Client, m *irc.Message) {
	go func(m *irc.Message) {
		configLock.RLock()
		defer configLock.RUnlock()

		if err := config.LogRawMessage(m); err != nil {
			log.WithError(err).Error("Unable to log raw message")
		}
	}(m)

	switch m.Command {
	case "001":
		// 001 is a welcome event, so we join channels there
		c.WriteMessage(&irc.Message{
			Command: "CAP",
			Params: []string{
				"REQ",
				strings.Join([]string{
					"twitch.tv/commands",
					"twitch.tv/membership",
					"twitch.tv/tags",
				}, " "),
			},
		})
		go i.ExecuteJoins(config.Channels)

	case "JOIN":
		// JOIN (Default IRC Command)
		// User enters the channel, might be triggered multiple times
		// should not be used to greet users
		i.handleJoin(m)

	case "NOTICE":
		// NOTICE (Twitch Commands)
		// General notices from the server.
		i.handleTwitchNotice(m)

	case "PART":
		// PART (Default IRC Command)
		// User leaves the channel, might be triggered multiple times
		i.handlePart(m)

	case "PRIVMSG":
		i.handleTwitchPrivmsg(m)

	case "RECONNECT":
		// RECONNECT (Twitch Commands)
		// In this case, reconnect and rejoin channels that were on the connection, as you would normally.
		log.Warn("We were asked to reconnect, closing connection")
		i.Close()

	case "USERNOTICE":
		// USERNOTICE (Twitch Commands)
		// Announces Twitch-specific events to the channel (for example, a user’s subscription notification).
		i.handleTwitchUsernotice(m)

	case "USERSTATE":
		// USERSTATE (Twitch Tags)
		// Sends user-state data when a user joins a channel or sends a PRIVMSG to a channel.
		i.handleTwitchUserstate(m)

	case "WHISPER":
		// WHISPER (Twitch Commands)
		// Delivers whisper-messages received
		i.handleTwitchWhisper(m)

	default:
		log.WithFields(log.Fields{
			"command":  m.Command,
			"tags":     m.Tags,
			"trailing": m.Trailing(),
		}).Trace("Unhandled message")
		// Unhandled message type, not yet needed
	}

	if err := notifyRawMessageHandlers(m); err != nil {
		log.WithError(err).Error("Unable to notify raw message handlers")
	}
}

func (i ircHandler) Run() error { return errors.Wrap(i.c.Run(), "running IRC client") }

func (i ircHandler) SendMessage(m *irc.Message) error { return i.c.WriteMessage(m) }

func (ircHandler) getChannel(m *irc.Message) string {
	if len(m.Params) > 0 {
		return m.Params[0]
	}
	return ""
}

func (i ircHandler) handleJoin(m *irc.Message) {
	go handleMessage(i.c, m, eventTypeJoin, nil)
}

func (i ircHandler) handlePart(m *irc.Message) {
	go handleMessage(i.c, m, eventTypePart, nil)
}

func (i ircHandler) handlePermit(m *irc.Message) {
	badges := twitch.ParseBadgeLevels(m)
	if !badges.Has(twitch.BadgeBroadcaster) && (!config.PermitAllowModerator || !badges.Has(twitch.BadgeModerator)) {
		// Neither broadcaster nor moderator or moderator not permitted
		return
	}

	msgParts := strings.Split(m.Trailing(), " ")
	if len(msgParts) != 2 { //nolint:gomnd // This is not a magic number but just an expected count
		return
	}

	username := msgParts[1]

	log.WithField("user", username).Debug("Added permit")
	timerStore.AddPermit(m.Params[0], username)

	go handleMessage(i.c, m, eventTypePermit, plugins.FieldCollection{"username": username})
}

func (i ircHandler) handleTwitchNotice(m *irc.Message) {
	log.WithFields(log.Fields{
		"channel":  i.getChannel(m),
		"tags":     m.Tags,
		"trailing": m.Trailing(),
	}).Trace("IRC NOTICE event")

	switch m.Tags["msg-id"] {
	case "":
		// Notices SHOULD have msg-id tags...
		log.WithField("msg", m).Warn("Received notice without msg-id")

	case "host_success", "host_success_viewers":
		log.WithField("trailing", m.Trailing()).Warn("Incoming host")

		go handleMessage(i.c, m, eventTypeHost, nil)

	}
}

func (i ircHandler) handleTwitchPrivmsg(m *irc.Message) {
	log.WithFields(log.Fields{
		"channel":  i.getChannel(m),
		"name":     m.Name,
		"user":     m.User,
		"tags":     m.Tags,
		"trailing": m.Trailing(),
	}).Trace("Received privmsg")

	if m.User != i.user {
		// Count messages from other users than self
		configLock.RLock()
		for _, am := range config.AutoMessages {
			am.CountMessage(m.Params[0])
		}
		configLock.RUnlock()
	}

	if strings.HasPrefix(m.Trailing(), "!permit") {
		i.handlePermit(m)
		return
	}

	if bits, err := strconv.ParseInt(string(m.Tags["bits"]), 10, 64); err == nil {
		go handleMessage(i.c, m, eventTypeBits, plugins.FieldCollection{
			"bits": bits,
		})
	}

	go handleMessage(i.c, m, nil, nil)
}

func (i ircHandler) handleTwitchUsernotice(m *irc.Message) {
	log.WithFields(log.Fields{
		"channel":  i.getChannel(m),
		"tags":     m.Tags,
		"trailing": m.Trailing(),
	}).Trace("IRC USERNOTICE event")

	switch m.Tags["msg-id"] {
	case "":
		// Notices SHOULD have msg-id tags...
		log.WithField("msg", m).Warn("Received usernotice without msg-id")

	case "raid":
		evtData := plugins.FieldCollection{
			"channel":     i.getChannel(m), // Compatibility to plugins.DeriveChannel
			"from":        m.Tags["login"],
			"user":        m.Tags["login"], // Compatibility to plugins.DeriveUser
			"viewercount": m.Tags["msg-param-viewerCount"],
		}
		log.WithFields(log.Fields(evtData)).Info("Incoming raid")

		go handleMessage(i.c, m, eventTypeRaid, evtData)

	case "resub":
		evtData := plugins.FieldCollection{
			"channel":           i.getChannel(m), // Compatibility to plugins.DeriveChannel
			"from":              m.Tags["login"],
			"subscribed_months": m.Tags["msg-param-cumulative-months"],
			"plan":              m.Tags["msg-param-sub-plan"],
			"user":              m.Tags["login"], // Compatibility to plugins.DeriveUser
		}
		log.WithFields(log.Fields(evtData)).Info("User re-subscribed")

		go handleMessage(i.c, m, eventTypeResub, evtData)

	case "sub":
		evtData := plugins.FieldCollection{
			"channel": i.getChannel(m), // Compatibility to plugins.DeriveChannel
			"from":    m.Tags["login"],
			"plan":    m.Tags["msg-param-sub-plan"],
			"user":    m.Tags["login"], // Compatibility to plugins.DeriveUser
		}
		log.WithFields(log.Fields(evtData)).Info("User subscribed")

		go handleMessage(i.c, m, eventTypeSub, evtData)

	case "subgift", "anonsubgift":
		evtData := plugins.FieldCollection{
			"channel":       i.getChannel(m), // Compatibility to plugins.DeriveChannel
			"from":          m.Tags["login"],
			"gifted_months": m.Tags["msg-param-gift-months"],
			"plan":          m.Tags["msg-param-sub-plan"],
			"to":            m.Tags["msg-param-recipient-user-name"],
			"user":          m.Tags["login"], // Compatibility to plugins.DeriveUser
		}
		log.WithFields(log.Fields(evtData)).Info("User gifted a sub")

		go handleMessage(i.c, m, eventTypeSubgift, evtData)

	case "submysterygift":
		evtData := plugins.FieldCollection{
			"channel": i.getChannel(m), // Compatibility to plugins.DeriveChannel
			"from":    m.Tags["login"],
			"number":  m.Tags["msg-param-mass-gift-count"],
			"plan":    m.Tags["msg-param-sub-plan"],
			"user":    m.Tags["login"], // Compatibility to plugins.DeriveUser
		}
		log.WithFields(log.Fields(evtData)).Info("User gifted subs to the community")

		go handleMessage(i.c, m, eventTypeSubmysterygift, evtData)

	}
}

func (i ircHandler) handleTwitchUserstate(m *irc.Message) {
	state, err := parseTwitchUserState(m)
	if err != nil {
		log.WithError(err).Error("Unable to parse bot user-state")
		return
	}

	botUserstate.Set(plugins.DeriveChannel(m, nil), state)
}

func (i ircHandler) handleTwitchWhisper(m *irc.Message) {
	go handleMessage(i.c, m, eventTypeWhisper, nil)
}
