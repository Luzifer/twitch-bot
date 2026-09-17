package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/pkg/event"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

type ircHandler struct {
	c           *irc.Client
	conn        *tls.Conn
	ctx         context.Context //nolint:containedctx // just stored internally
	ctxCancelFn func()
	user        string
}

var (
	rawMessageHandlers     []plugins.RawMessageHandlerFunc
	rawMessageHandlersLock sync.Mutex
)

func notifyRawMessageHandlers(m *irc.Message) error {
	rawMessageHandlersLock.Lock()
	defer rawMessageHandlersLock.Unlock()

	for _, fn := range rawMessageHandlers {
		if err := fn(m); err != nil {
			return fmt.Errorf("executing raw message handlers: %w", err)
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

func newIRCHandler() (*ircHandler, error) {
	h := new(ircHandler)

	_, username, err := twitchClient.GetAuthorizedUser(context.Background())
	if err != nil {
		return nil, fmt.Errorf("fetching username: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background()) //#nosec:G118 // Cancel is retained in the handler and called on constructor failure or handler shutdown
	h.ctx, h.ctxCancelFn = ctx, cancel

	handlerReady := false
	defer func() {
		if !handlerReady {
			cancel()
		}
	}()

	conn, err := tls.Dial("tcp", "irc.chat.twitch.tv:6697", nil) //nolint:noctx // Would use background context
	if err != nil {
		return nil, fmt.Errorf("connect to IRC server: %w", err)
	}

	token, err := twitchClient.GetToken(context.Background())
	if err != nil {
		return nil, fmt.Errorf("getting auth token: %w", err)
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

	handlerReady = true
	return h, nil
}

func (i ircHandler) Client() *irc.Client { return i.c }

func (i ircHandler) Close() error {
	i.ctxCancelFn()
	return nil
}

func (i ircHandler) ExecuteJoins(channels []string) {
	for _, ch := range channels {
		_ = i.c.Write(fmt.Sprintf("JOIN #%s", strings.TrimLeft(ch, "#")))
	}
}

func (i ircHandler) ExecutePart(channel string) {
	_ = i.c.Write(fmt.Sprintf("PART #%s", strings.TrimLeft(channel, "#")))
}

//nolint:gocyclo // this is only a simple distribution-list without much logic
func (i ircHandler) Handle(c *irc.Client, m *irc.Message) {
	// We've received a message, update status check
	statusIRCMessageReceived = time.Now()

	go func(m *irc.Message) {
		configLock.RLock()
		defer configLock.RUnlock()

		if err := config.LogRawMessage(m); err != nil {
			logrus.WithError(err).Error("Unable to log raw message")
		}
	}(m)

	if m.Tags["source-room-id"] != "" && m.Tags["source-room-id"] != m.Tags["room-id"] {
		// Message has its `source-room-id` set, which signals it
		// originates from a shared chat. Additionally its `source-room-id`
		// does not match the `room-id` which we are listening to. So we
		// shouldn't care about handling that message in order to prevent
		// false bit alerts or reactions to `!so` or other "shared" commands.
		return
	}

	switch m.Command {
	case "001":
		// 001 is a welcome event, so we join channels there
		_ = c.WriteMessage(&irc.Message{
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
		logrus.Warn("We were asked to reconnect, closing connection")
		if err := i.Close(); err != nil {
			logrus.WithError(err).Error("closing IRC connection after reconnect")
		}

	case "USERNOTICE":
		// USERNOTICE (Twitch Commands)
		// Announces Twitch-specific events to the channel (for example, a user’s subscription notification).
		i.handleTwitchUsernotice(m)

	case "WHISPER":
		// WHISPER (Twitch Commands)
		// Delivers whisper-messages received
		i.handleTwitchWhisper(m)

	default:
		logrus.WithFields(logrus.Fields{
			"command":  m.Command,
			"tags":     m.Tags,
			"trailing": m.Trailing(),
		}).Trace("Unhandled message")
		// Unhandled message type, not yet needed
	}

	if err := notifyRawMessageHandlers(m); err != nil {
		logrus.WithError(err).Error("Unable to notify raw message handlers")
	}
}

func (i ircHandler) Run() error {
	if err := i.c.RunContext(i.ctx); err != nil {
		return fmt.Errorf("running IRC client: %w", err)
	}
	return nil
}

func (i ircHandler) SendMessage(m *irc.Message) (err error) {
	if err = i.c.WriteMessage(m); err != nil {
		return fmt.Errorf("writing message: %w", err)
	}
	return nil
}

func (ircHandler) getChannel(m *irc.Message) string {
	if len(m.Params) > 0 {
		return m.Params[0]
	}
	return ""
}

func (i ircHandler) handleClearChat(m *irc.Message) {
	seconds, secondsErr := strconv.Atoi(m.Tags["ban-duration"])
	targetUserID, hasTargetUserID := m.Tags["target-user-id"]
	channel := i.getChannel(m) // Compatibility to plugins.DeriveChannel

	var evt event.Event

	switch {
	case secondsErr == nil && hasTargetUserID:
		// User & Duration = Timeout
		evt = event.Timeout{
			Channel:    channel,
			Duration:   time.Duration(seconds) * time.Second,
			Seconds:    seconds,
			TargetID:   targetUserID,
			TargetName: m.Trailing(),
		}

		logrus.WithFields(event.ToLogFields(evt)).Info("User was timed out")

	case hasTargetUserID:
		// User w/o Duration = Ban
		evt = event.Ban{
			Channel:    channel,
			TargetID:   targetUserID,
			TargetName: m.Trailing(),
		}

		logrus.WithFields(event.ToLogFields(evt)).Info("User was banned")

	default:
		// No User = /clear
		evt = event.ClearChat{
			Channel: channel,
		}

		logrus.WithFields(event.ToLogFields(evt)).Info("Chat was cleared")
	}

	go handleTypedMessage(i.c, m, evt)
}

func (i ircHandler) handleClearMessage(m *irc.Message) {
	evt := event.ClearMessage{
		Channel:    i.getChannel(m), // Compatibility to plugins.DeriveChannel
		MessageID:  m.Tags["target-msg-id"],
		TargetName: m.Tags["login"],
	}

	logrus.WithFields(event.ToLogFields(evt)).
		WithField("message", m.Trailing()).
		Info("Message was deleted")

	go handleTypedMessage(i.c, m, evt)
}

func (i ircHandler) handleJoin(m *irc.Message) {
	go handleTypedMessage(i.c, m, event.Join{
		Channel: i.getChannel(m), // Compatibility to plugins.DeriveChannel
		User:    m.User,          // Compatibility to plugins.DeriveUser
	})
}

func (i ircHandler) handlePart(m *irc.Message) {
	go handleTypedMessage(i.c, m, event.Part{
		Channel: i.getChannel(m), // Compatibility to plugins.DeriveChannel
		User:    m.User,          // Compatibility to plugins.DeriveUser
	})
}

func (i ircHandler) handlePermit(m *irc.Message) {
	badges := twitch.ParseBadgeLevels(m)
	if !badges.Has(twitch.BadgeBroadcaster) && (!config.PermitAllowModerator || !badges.Has(twitch.BadgeModerator)) {
		// Neither broadcaster nor moderator or moderator not permitted
		return
	}

	msgParts := strings.Split(m.Trailing(), " ")
	if len(msgParts) != 2 { //nolint:mnd // This is not a magic number but just an expected count
		return
	}

	username := msgParts[1]

	evt := event.Permit{
		Channel: i.getChannel(m), // Compatibility to plugins.DeriveChannel
		To:      username,
		User:    m.User, // Compatibility to plugins.DeriveUser
		UserID:  m.Tags["user-id"],
	}

	logrus.WithFields(event.ToLogFields(evt)).Debug("Added permit")
	if err := timerService.AddPermit(m.Params[0], username); err != nil {
		logrus.WithError(err).Error("adding permit")
	}

	go handleTypedMessage(i.c, m, evt)
}

func (i ircHandler) handleTwitchNotice(m *irc.Message) {
	logrus.WithFields(logrus.Fields{
		eventFieldChannel: i.getChannel(m),
		"tags":            m.Tags,
		"trailing":        m.Trailing(),
	}).Trace("IRC NOTICE event")

	switch m.Tags["msg-id"] {
	case "":
		// Notices SHOULD have msg-id tags...
		logrus.WithField("msg", m).Warn("Received notice without msg-id")

	default:
		logrus.WithField("id", m.Tags["msg-id"]).Debug("unhandled notice received")
	}
}

func (i ircHandler) handleTwitchPrivmsg(m *irc.Message) {
	logrus.WithFields(logrus.Fields{
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
		evt := event.Bits{
			Bits:    bits,
			Channel: i.getChannel(m), // Compatibility to plugins.DeriveChannel
			Message: m.Trailing(),
			User:    m.User, // Compatibility to plugins.DeriveUser
			UserID:  m.Tags["user-id"],
		}

		logrus.WithFields(event.ToLogFields(evt)).Info("User spent bits in chat message")

		go handleTypedMessage(i.c, m, evt)
	}

	go handleTypedMessage(i.c, m, event.Message{})
}

//nolint:funlen // just a list of mappings
func (i ircHandler) handleTwitchUsernotice(m *irc.Message) {
	logrus.WithFields(logrus.Fields{
		eventFieldChannel: i.getChannel(m),
		"tags":            m.Tags,
		"trailing":        m.Trailing(),
	}).Trace("IRC USERNOTICE event")

	var (
		baseChannel = event.BaseChannel{
			Channel: i.getChannel(m), // Compatibility to plugins.DeriveChannel
		}
		baseUser = event.BaseUser{
			User:   m.Tags["login"], // Compatibility to plugins.DeriveUser
			UserID: m.Tags["user-id"],
		}
	)

	message := m.Trailing()
	if message == i.getChannel(m) {
		// If no message is given, Trailing yields the channel name
		message = ""
	}

	switch m.Tags["msg-id"] {
	case "":
		// Notices SHOULD have msg-id tags...
		logrus.WithField("msg", m).Warn("Received usernotice without msg-id")

	case "announcement":
		evt := event.Announcement{
			BaseChannel: baseChannel,
			BaseUser:    baseUser,
			Color:       m.Tags["msg-param-color"],
			Message:     m.Trailing(),
		}

		logrus.WithFields(event.ToLogFields(evt)).Info("Announcement was made")
		go handleTypedMessage(i.c, m, evt)

	case "giftpaidupgrade":
		evt := event.GiftPaidUpgrade{
			BaseChannel: baseChannel,
			BaseUser:    baseUser,
			Gifter:      m.Tags["msg-param-sender-login"],
		}

		logrus.WithFields(event.ToLogFields(evt)).Info("User upgraded from gift to paid sub")
		go handleTypedMessage(i.c, m, evt)

	case "primepaidupgrade":
		evt := event.PrimePaidUpgrade{
			BaseChannel: baseChannel,
			BaseUser:    baseUser,
			Plan:        m.Tags["msg-param-sub-plan"],
		}

		logrus.WithFields(event.ToLogFields(evt)).Info("User upgraded from prime to paid sub")
		go handleTypedMessage(i.c, m, evt)

	case "raid":
		evt := event.Raid{
			BaseChannel: baseChannel,
			BaseUser:    baseUser,
			From:        m.Tags["login"],
			ViewerCount: i.tagToNumeric(m, "msg-param-viewerCount", 0),
		}

		logrus.WithFields(event.ToLogFields(evt)).Info("Incoming raid")
		go handleTypedMessage(i.c, m, evt)

	case "resub":
		evt := event.Resub{
			BaseChannel:      baseChannel,
			BaseUser:         baseUser,
			From:             m.Tags["login"],
			Message:          message,
			MultiMonth:       i.tagToNumeric(m, "msg-param-multimonth-duration", 0),
			Plan:             m.Tags["msg-param-sub-plan"],
			SubscribedMonths: i.tagToNumeric(m, "msg-param-cumulative-months", 0),
		}

		logrus.WithFields(event.ToLogFields(evt)).Info("User re-subscribed")
		go handleTypedMessage(i.c, m, evt)

	case "sub":
		evt := event.Sub{
			BaseChannel: baseChannel,
			BaseUser:    baseUser,
			From:        m.Tags["login"],
			MultiMonth:  i.tagToNumeric(m, "msg-param-multimonth-duration", 0),
			Plan:        m.Tags["msg-param-sub-plan"],
		}

		logrus.WithFields(event.ToLogFields(evt)).Info("User subscribed")
		go handleTypedMessage(i.c, m, evt)

	case "subgift", "anonsubgift":
		evt := event.SubGift{
			BaseChannel:      baseChannel,
			BaseUser:         baseUser,
			From:             m.Tags["login"],
			GiftedMonths:     i.tagToNumeric(m, "msg-param-gift-months", 1),
			MultiMonth:       i.tagToNumeric(m, "msg-param-multimonth-duration", 0),
			OriginID:         m.Tags["msg-param-origin-id"],
			Plan:             m.Tags["msg-param-sub-plan"],
			SubscribedMonths: i.tagToNumeric(m, "msg-param-months", 0),
			To:               m.Tags["msg-param-recipient-user-name"],
			TotalGifted:      i.tagToNumeric(m, "msg-param-sender-count", 0),
		}

		logrus.WithFields(event.ToLogFields(evt)).Info("User gifted a sub")
		go handleTypedMessage(i.c, m, evt)

	case "submysterygift":
		evt := event.SubMysteryGift{
			BaseChannel: baseChannel,
			BaseUser:    baseUser,
			From:        m.Tags["login"],
			MultiMonth:  i.tagToNumeric(m, "msg-param-multimonth-duration", 0),
			Number:      i.tagToNumeric(m, "msg-param-mass-gift-count", 0),
			OriginID:    m.Tags["msg-param-origin-id"],
			Plan:        m.Tags["msg-param-sub-plan"],
			TotalGifted: i.tagToNumeric(m, "msg-param-sender-count", 0),
		}

		logrus.WithFields(event.ToLogFields(evt)).Info("User gifted subs to the community")
		go handleTypedMessage(i.c, m, evt)

	case "viewermilestone":
		switch m.Tags["msg-param-category"] {
		case "watch-streak":
			evt := event.WatchStreak{
				BaseChannel: baseChannel,
				BaseUser:    baseUser,
				Message:     message,
				Streak:      i.tagToNumeric(m, "msg-param-value", 0),
			}

			logrus.WithFields(event.ToLogFields(evt)).Info("User shared a watch-streak")
			go handleTypedMessage(i.c, m, evt)

		default:
			logrus.WithField("category", m.Tags["msg-param-category"]).Debug("found unhandled viewermilestone category")
		}
	}
}

func (i ircHandler) handleTwitchWhisper(m *irc.Message) {
	go handleTypedMessage(i.c, m, event.Whisper{})
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
