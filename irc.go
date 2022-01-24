package main

import (
	"crypto/tls"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/Luzifer/twitch-bot/twitch"
)

var (
	rawMessageHandlers     []plugins.RawMessageHandlerFunc
	rawMessageHandlersLock sync.Mutex

	hostNotificationRegex = regexp.MustCompile(`^(?P<actor>\w+) is now(?: auto)? hosting you(?: for(?: up to)? (?P<amount>[0-9]+) viewers)?.$`)
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

	token, err := twitchClient.GetToken()
	if err != nil {
		return nil, errors.Wrap(err, "getting auth token")
	}

	h.c = irc.NewClient(conn, irc.ClientConfig{
		Nick:    username,
		Pass:    strings.Join([]string{"oauth", token}, ":"),
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
	// We've received a message, update status check
	statusIRCMessageReceived = time.Now()

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

	case "CLEARCHAT":
		// CLEARCHAT (Twitch Commands)
		// Purge a user’s messages, typically after a user is banned from
		// chat or timed out.
		i.handleClearChat(m)

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

	case "PING":
		// PING (Default IRC Command)
		// Handled by the library, just here to prevent trace-logging every ping

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

func (i ircHandler) handleClearChat(m *irc.Message) {
	seconds, secondsErr := strconv.Atoi(string(m.Tags["ban-duration"]))
	targetUserID, hasTargetUserID := m.Tags.GetTag("target-user-id")

	var (
		evt    *string
		fields = plugins.NewFieldCollection()
	)

	fields.Set("channel", i.getChannel(m)) // Compatibility to plugins.DeriveChannel

	switch {
	case secondsErr == nil && hasTargetUserID:
		// User & Duration = Timeout
		evt = eventTypeTimeout
		fields.Set("duration", time.Duration(seconds)*time.Second)
		fields.Set("seconds", seconds)
		fields.Set("target_id", targetUserID)
		fields.Set("target_name", m.Trailing())
		log.WithFields(log.Fields(fields.Data())).Info("User was timed out")

	case hasTargetUserID:
		// User w/o Duration = Ban
		evt = eventTypeBan
		fields.Set("target_id", targetUserID)
		fields.Set("target_name", m.Trailing())
		log.WithFields(log.Fields(fields.Data())).Info("User was banned")

	default:
		// No User = /clear
		evt = eventTypeClearChat
		log.WithFields(log.Fields(fields.Data())).Info("Chat was cleared")
	}

	go handleMessage(i.c, m, evt, fields)
}

func (i ircHandler) handleJoin(m *irc.Message) {
	fields := plugins.FieldCollectionFromData(map[string]interface{}{
		"channel": i.getChannel(m), // Compatibility to plugins.DeriveChannel
		"user":    m.User,          // Compatibility to plugins.DeriveUser
	})
	go handleMessage(i.c, m, eventTypeJoin, fields)
}

func (i ircHandler) handlePart(m *irc.Message) {
	fields := plugins.FieldCollectionFromData(map[string]interface{}{
		"channel": i.getChannel(m), // Compatibility to plugins.DeriveChannel
		"user":    m.User,          // Compatibility to plugins.DeriveUser
	})
	go handleMessage(i.c, m, eventTypePart, fields)
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

	fields := plugins.FieldCollectionFromData(map[string]interface{}{
		"channel":  i.getChannel(m), // Compatibility to plugins.DeriveChannel
		"user":     m.User,          // Compatibility to plugins.DeriveUser
		"username": username,        // DEPRECATED but kept for comapatibility
		"to":       username,
	})

	log.WithFields(fields.Data()).Debug("Added permit")
	timerStore.AddPermit(m.Params[0], username)

	go handleMessage(i.c, m, eventTypePermit, fields)
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

		fields := plugins.FieldCollectionFromData(map[string]interface{}{
			"channel": i.getChannel(m), // Compatibility to plugins.DeriveChannel
			"user":    m.User,          // Compatibility to plugins.DeriveUser
		})
		go handleMessage(i.c, m, eventTypeHost, fields)

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

	// Handle the jtv host message for hosts
	if m.User == "jtv" && hostNotificationRegex.MatchString(m.Trailing()) {
		matches := hostNotificationRegex.FindStringSubmatch(m.Trailing())
		if matches[2] == "" {
			matches[2] = "0"
		}

		fields := plugins.FieldCollectionFromData(map[string]interface{}{
			"channel":     fmt.Sprintf("#%s", i.user),
			"from":        matches[1],
			"viewerCount": 0,
		})

		if v, err := strconv.Atoi(matches[2]); err == nil {
			fields.Set("viewerCount", v)
		}

		log.WithFields(log.Fields(fields.Data())).Info("Incoming Host (jtv announce)")

		go handleMessage(i.c, m, eventTypeHost, fields)
		return
	}

	if strings.HasPrefix(m.Trailing(), "!permit") {
		i.handlePermit(m)
		return
	}

	if bits, err := strconv.ParseInt(string(m.Tags["bits"]), 10, 64); err == nil {
		fields := plugins.FieldCollectionFromData(map[string]interface{}{
			"bits":    bits,
			"channel": i.getChannel(m), // Compatibility to plugins.DeriveChannel
			"user":    m.User,          // Compatibility to plugins.DeriveUser
		})

		log.WithFields(log.Fields(fields.Data())).Info("User spent bits in chat message")

		go handleMessage(i.c, m, eventTypeBits, fields)
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

	case "giftpaidupgrade":
		evtData := plugins.FieldCollectionFromData(map[string]interface{}{
			"channel": i.getChannel(m), // Compatibility to plugins.DeriveChannel
			"gifter":  m.Tags["msg-param-sender-login"],
			"user":    m.Tags["login"], // Compatibility to plugins.DeriveUser
		})
		log.WithFields(log.Fields(evtData.Data())).Info("User upgraded to paid sub")

		go handleMessage(i.c, m, eventTypeGiftPaidUpgrade, evtData)

	case "raid":
		evtData := plugins.FieldCollectionFromData(map[string]interface{}{
			"channel":     i.getChannel(m), // Compatibility to plugins.DeriveChannel
			"from":        m.Tags["login"],
			"user":        m.Tags["login"], // Compatibility to plugins.DeriveUser
			"viewercount": m.Tags["msg-param-viewerCount"],
		})
		log.WithFields(log.Fields(evtData.Data())).Info("Incoming raid")

		go handleMessage(i.c, m, eventTypeRaid, evtData)

	case "resub":
		evtData := plugins.FieldCollectionFromData(map[string]interface{}{
			"channel":           i.getChannel(m), // Compatibility to plugins.DeriveChannel
			"from":              m.Tags["login"],
			"subscribed_months": m.Tags["msg-param-cumulative-months"],
			"plan":              m.Tags["msg-param-sub-plan"],
			"user":              m.Tags["login"], // Compatibility to plugins.DeriveUser
		})
		log.WithFields(log.Fields(evtData.Data())).Info("User re-subscribed")

		go handleMessage(i.c, m, eventTypeResub, evtData)

	case "sub":
		evtData := plugins.FieldCollectionFromData(map[string]interface{}{
			"channel": i.getChannel(m), // Compatibility to plugins.DeriveChannel
			"from":    m.Tags["login"],
			"plan":    m.Tags["msg-param-sub-plan"],
			"user":    m.Tags["login"], // Compatibility to plugins.DeriveUser
		})
		log.WithFields(log.Fields(evtData.Data())).Info("User subscribed")

		go handleMessage(i.c, m, eventTypeSub, evtData)

	case "subgift", "anonsubgift":
		evtData := plugins.FieldCollectionFromData(map[string]interface{}{
			"channel":       i.getChannel(m), // Compatibility to plugins.DeriveChannel
			"from":          m.Tags["login"],
			"gifted_months": m.Tags["msg-param-gift-months"],
			"plan":          m.Tags["msg-param-sub-plan"],
			"to":            m.Tags["msg-param-recipient-user-name"],
			"user":          m.Tags["login"], // Compatibility to plugins.DeriveUser
		})
		log.WithFields(log.Fields(evtData.Data())).Info("User gifted a sub")

		go handleMessage(i.c, m, eventTypeSubgift, evtData)

	case "submysterygift":
		evtData := plugins.FieldCollectionFromData(map[string]interface{}{
			"channel": i.getChannel(m), // Compatibility to plugins.DeriveChannel
			"from":    m.Tags["login"],
			"number":  m.Tags["msg-param-mass-gift-count"],
			"plan":    m.Tags["msg-param-sub-plan"],
			"user":    m.Tags["login"], // Compatibility to plugins.DeriveUser
		})
		log.WithFields(log.Fields(evtData.Data())).Info("User gifted subs to the community")

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
