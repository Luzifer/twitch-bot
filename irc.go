package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
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
	c           *irc.Client
	conn        *tls.Conn
	ctx         context.Context
	ctxCancelFn func()
	user        string
}

func newIRCHandler() (*ircHandler, error) {
	h := new(ircHandler)

	_, username, err := twitchClient.GetAuthorizedUser()
	if err != nil {
		return nil, errors.Wrap(err, "fetching username")
	}

	h.ctx, h.ctxCancelFn = context.WithCancel(context.Background())

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

func (i ircHandler) Close() error {
	i.ctxCancelFn()
	return nil
}

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

	case "CLEARMSG":
		// CLEARMSG (Twitch Commands)
		// Removes a single message from a channel. This is triggered by
		// the/delete <target-msg-id> command on IRC.
		i.handleClearMessage(m)

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

func (i ircHandler) Run() error { return errors.Wrap(i.c.RunContext(i.ctx), "running IRC client") }

func (i ircHandler) SendMessage(m *irc.Message) error { return i.c.WriteMessage(m) }

func (ircHandler) getChannel(m *irc.Message) string {
	if len(m.Params) > 0 {
		return m.Params[0]
	}
	return ""
}

func (i ircHandler) handleClearChat(m *irc.Message) {
	seconds, secondsErr := strconv.Atoi(m.Tags["ban-duration"])
	targetUserID, hasTargetUserID := m.Tags["target-user-id"]

	var (
		evt    *string
		fields = plugins.NewFieldCollection()
	)

	fields.Set(eventFieldChannel, i.getChannel(m)) // Compatibility to plugins.DeriveChannel

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

func (i ircHandler) handleClearMessage(m *irc.Message) {
	fields := plugins.FieldCollectionFromData(map[string]interface{}{
		eventFieldChannel: i.getChannel(m), // Compatibility to plugins.DeriveChannel
		"message_id":      m.Tags["target-msg-id"],
		"target_name":     m.Tags["login"],
	})
	log.WithFields(log.Fields(fields.Data())).
		WithField("message", m.Trailing()).
		Info("Message was deleted")
	go handleMessage(i.c, m, eventTypeDelete, fields)
}

func (i ircHandler) handleJoin(m *irc.Message) {
	fields := plugins.FieldCollectionFromData(map[string]interface{}{
		eventFieldChannel:  i.getChannel(m), // Compatibility to plugins.DeriveChannel
		eventFieldUserName: m.User,          // Compatibility to plugins.DeriveUser
	})
	go handleMessage(i.c, m, eventTypeJoin, fields)
}

func (i ircHandler) handlePart(m *irc.Message) {
	fields := plugins.FieldCollectionFromData(map[string]interface{}{
		eventFieldChannel:  i.getChannel(m), // Compatibility to plugins.DeriveChannel
		eventFieldUserName: m.User,          // Compatibility to plugins.DeriveUser
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
		eventFieldChannel:  i.getChannel(m), // Compatibility to plugins.DeriveChannel
		eventFieldUserName: m.User,          // Compatibility to plugins.DeriveUser
		eventFieldUserID:   m.Tags["user-id"],
		"to":               username,
	})

	log.WithFields(fields.Data()).Debug("Added permit")
	timerService.AddPermit(m.Params[0], username)

	go handleMessage(i.c, m, eventTypePermit, fields)
}

func (i ircHandler) handleTwitchNotice(m *irc.Message) {
	log.WithFields(log.Fields{
		eventFieldChannel: i.getChannel(m),
		"tags":            m.Tags,
		"trailing":        m.Trailing(),
	}).Trace("IRC NOTICE event")

	switch m.Tags["msg-id"] {
	case "":
		// Notices SHOULD have msg-id tags...
		log.WithField("msg", m).Warn("Received notice without msg-id")

	default:
		log.WithField("id", m.Tags["msg-id"]).Debug("unhandled notice received")
	}
}

func (i ircHandler) handleTwitchPrivmsg(m *irc.Message) {
	log.WithFields(log.Fields{
		eventFieldChannel:  i.getChannel(m),
		"name":             m.Name,
		eventFieldUserName: m.User,
		eventFieldUserID:   m.Tags["user-id"],
		"tags":             m.Tags,
		"trailing":         m.Trailing(),
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

	if bits := i.tagToNumeric(m, "bits", 0); bits > 0 {
		fields := plugins.FieldCollectionFromData(map[string]interface{}{
			"bits":             bits,
			eventFieldChannel:  i.getChannel(m), // Compatibility to plugins.DeriveChannel
			"message":          m.Trailing(),
			eventFieldUserName: m.User, // Compatibility to plugins.DeriveUser
			eventFieldUserID:   m.Tags["user-id"],
		})

		log.WithFields(log.Fields(fields.Data())).Info("User spent bits in chat message")

		go handleMessage(i.c, m, eventTypeBits, fields)
	}

	if amount := i.tagToNumeric(m, "pinned-chat-paid-amount", 0); amount > 0 {
		fields := plugins.FieldCollectionFromData(map[string]any{
			"amount":            float64(amount) / math.Pow10(int(i.tagToNumeric(m, "pinned-chat-paid-exponent", 0))),
			"currency":          m.Tags["pinned-chat-paid-currency"],
			eventFieldChannel:   i.getChannel(m),
			eventFieldUserID:    m.Tags["user-id"],
			eventFieldUserName:  m.User,
			"is_system_message": m.Tags["pinned-chat-paid-is-system-message"] == "1",
			"level":             m.Tags["pinned-chat-paid-level"],
			"message":           m.Trailing(),
		})

		log.WithFields(log.Fields(fields.Data())).Info("User used hype-chat message")

		go handleMessage(i.c, m, eventTypeHypeChat, fields)
	}

	go handleMessage(i.c, m, nil, nil)
}

func (i ircHandler) handleTwitchUsernotice(m *irc.Message) {
	log.WithFields(log.Fields{
		eventFieldChannel: i.getChannel(m),
		"tags":            m.Tags,
		"trailing":        m.Trailing(),
	}).Trace("IRC USERNOTICE event")

	evtData := plugins.FieldCollectionFromData(map[string]any{
		eventFieldChannel:  i.getChannel(m), // Compatibility to plugins.DeriveChannel
		eventFieldUserName: m.Tags["login"], // Compatibility to plugins.DeriveUser
		eventFieldUserID:   m.Tags["user-id"],
	})

	switch m.Tags["msg-id"] {
	case "":
		// Notices SHOULD have msg-id tags...
		log.WithField("msg", m).Warn("Received usernotice without msg-id")

	case "announcement":
		evtData.SetFromData(map[string]any{
			"color":   m.Tags["msg-param-color"],
			"message": m.Trailing(),
		})
		log.WithFields(log.Fields(evtData.Data())).Info("Announcement was made")

		go handleMessage(i.c, m, eventTypeAnnouncement, evtData)

	case "giftpaidupgrade":
		evtData.SetFromData(map[string]interface{}{
			"gifter": m.Tags["msg-param-sender-login"],
		})
		log.WithFields(log.Fields(evtData.Data())).Info("User upgraded to paid sub")

		go handleMessage(i.c, m, eventTypeGiftPaidUpgrade, evtData)

	case "raid":
		evtData.SetFromData(map[string]interface{}{
			"from":        m.Tags["login"],
			"viewercount": i.tagToNumeric(m, "msg-param-viewerCount", 0),
		})
		log.WithFields(log.Fields(evtData.Data())).Info("Incoming raid")

		go handleMessage(i.c, m, eventTypeRaid, evtData)

	case "resub":
		message := m.Trailing()
		if message == i.getChannel(m) {
			// If no message is given, Trailing yields the channel name
			message = ""
		}

		evtData.SetFromData(map[string]interface{}{
			"from":              m.Tags["login"],
			"message":           message,
			"multi_month":       i.tagToNumeric(m, "msg-param-multimonth-duration", 0),
			"subscribed_months": i.tagToNumeric(m, "msg-param-cumulative-months", 0),
			"plan":              m.Tags["msg-param-sub-plan"],
		})
		log.WithFields(log.Fields(evtData.Data())).Info("User re-subscribed")

		go handleMessage(i.c, m, eventTypeResub, evtData)

	case "sub":
		evtData.SetFromData(map[string]interface{}{
			"from":        m.Tags["login"],
			"multi_month": i.tagToNumeric(m, "msg-param-multimonth-duration", 0),
			"plan":        m.Tags["msg-param-sub-plan"],
		})
		log.WithFields(log.Fields(evtData.Data())).Info("User subscribed")

		go handleMessage(i.c, m, eventTypeSub, evtData)

	case "subgift", "anonsubgift":
		evtData.SetFromData(map[string]interface{}{
			"from":              m.Tags["login"],
			"gifted_months":     i.tagToNumeric(m, "msg-param-gift-months", 1),
			"multi_month":       i.tagToNumeric(m, "msg-param-multimonth-duration", 0),
			"origin_id":         m.Tags["msg-param-origin-id"],
			"plan":              m.Tags["msg-param-sub-plan"],
			"subscribed_months": i.tagToNumeric(m, "msg-param-months", 0),
			"to":                m.Tags["msg-param-recipient-user-name"],
			"total_gifted":      i.tagToNumeric(m, "msg-param-sender-count", 0),
		})
		log.WithFields(log.Fields(evtData.Data())).Info("User gifted a sub")

		go handleMessage(i.c, m, eventTypeSubgift, evtData)

	case "submysterygift":
		evtData.SetFromData(map[string]interface{}{
			"from":         m.Tags["login"],
			"multi_month":  i.tagToNumeric(m, "msg-param-multimonth-duration", 0),
			"number":       i.tagToNumeric(m, "msg-param-mass-gift-count", 0),
			"origin_id":    m.Tags["msg-param-origin-id"],
			"plan":         m.Tags["msg-param-sub-plan"],
			"total_gifted": i.tagToNumeric(m, "msg-param-sender-count", 0),
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

func (ircHandler) tagToNumeric(m *irc.Message, tag string, fallback int64) int64 {
	tv := m.Tags[tag]
	if tv == "" {
		return fallback
	}

	v, err := strconv.ParseInt(tv, 10, 64)
	if err != nil {
		return fallback
	}

	return v
}
