package main

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	badgeBroadcaster = "broadcaster"
	badgeFounder     = "founder"
	badgeModerator   = "moderator"
	badgeSubscriber  = "subscriber"
)

type ircHandler struct {
	conn *tls.Conn
	c    *irc.Client
	user string
}

func newIRCHandler() (*ircHandler, error) {
	h := new(ircHandler)

	username, err := twitch.getAuthorizedUsername()
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

func (i ircHandler) Close() error { return i.conn.Close() }

func (i ircHandler) ExecuteJoins(channels []string) {
	for _, ch := range channels {
		i.c.Write(fmt.Sprintf("JOIN #%s", strings.TrimLeft(ch, "#")))
	}
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
}

func (i ircHandler) Run() error { return errors.Wrap(i.c.Run(), "running IRC client") }

func (ircHandler) getChannel(m *irc.Message) string {
	if len(m.Params) > 0 {
		return m.Params[0]
	}
	return ""
}

func (i ircHandler) handleJoin(m *irc.Message) {
	go handleMessage(i.c, m, eventTypeJoin)
}

func (i ircHandler) handlePart(m *irc.Message) {
	go handleMessage(i.c, m, eventTypePart)
}

func (i ircHandler) handlePermit(m *irc.Message) {
	badges := i.ParseBadgeLevels(m)
	if !badges.Has(badgeBroadcaster) && (!config.PermitAllowModerator || !badges.Has(badgeModerator)) {
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

	go handleMessage(i.c, m, eventTypePermit)
}

func (i ircHandler) handleTwitchNotice(m *irc.Message) {
	log.WithFields(log.Fields{
		"channel":  i.getChannel(m),
		"tags":     m.Tags,
		"trailing": m.Trailing(),
	}).Debug("IRC NOTICE event")

	switch m.Tags["msg-id"] {
	case "":
		// Notices SHOULD have msg-id tags...
		log.WithField("msg", m).Warn("Received notice without msg-id")

	case "host_success", "host_success_viewers":
		log.WithField("trailing", m.Trailing()).Warn("Incoming host")

		go handleMessage(i.c, m, eventTypeHost)

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

	go handleMessage(i.c, m, nil)
}

func (i ircHandler) handleTwitchUsernotice(m *irc.Message) {
	log.WithFields(log.Fields{
		"channel":  i.getChannel(m),
		"tags":     m.Tags,
		"trailing": m.Trailing(),
	}).Debug("IRC USERNOTICE event")

	switch m.Tags["msg-id"] {
	case "":
		// Notices SHOULD have msg-id tags...
		log.WithField("msg", m).Warn("Received usernotice without msg-id")

	case "raid":
		log.WithFields(log.Fields{
			"from":        m.Tags["login"],
			"viewercount": m.Tags["msg-param-viewerCount"],
		}).Info("Incoming raid")

		go handleMessage(i.c, m, eventTypeRaid)

	case "resub":
		go handleMessage(i.c, m, eventTypeResub)

	case "sub":
		go handleMessage(i.c, m, eventTypeSub)

	case "subgift", "anonsubgift":
		go handleMessage(i.c, m, eventTypeSubgift)

	}
}

func (i ircHandler) handleTwitchWhisper(m *irc.Message) {
	go handleMessage(i.c, m, eventTypeWhisper)
}

func (ircHandler) ParseBadgeLevels(m *irc.Message) badgeCollection {
	out := badgeCollection{}

	badgeString, ok := m.GetTag("badges")
	if !ok || len(badgeString) == 0 {
		return out
	}

	badges := strings.Split(badgeString, ",")
	for _, b := range badges {
		badgeParts := strings.Split(b, "/")
		if len(badgeParts) != 2 { //nolint:gomnd // This is not a magic number but just an expected count
			log.WithField("badge", b).Warn("Malformed badge found")
			continue
		}

		level, err := strconv.Atoi(badgeParts[1])
		if err != nil {
			log.WithField("badge", b).Warn("Unparsable level in badge")
			continue
		}

		out.Add(badgeParts[0], level)
	}

	// If there is a founders badge but no subscribers badge
	// add a level-0 subscribers badge to prevent the bot to
	// cause trouble on founders when subscribers are allowed
	// to do something
	if out.Has(badgeFounder) && !out.Has(badgeSubscriber) {
		out.Add(badgeSubscriber, out.Get(badgeFounder))
	}

	return out
}
