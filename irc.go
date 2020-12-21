package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const twitchRequestTimeout = 2 * time.Second

const (
	badgeBroadcaster = "broadcaster"
	badgeFounder     = "founder"
	badgeModerator   = "moderator"
	badgeSubscriber  = "subscriber"
	badgeVIP         = "vip"
)

type ircHandler struct {
	conn *tls.Conn
	c    *irc.Client
	user string
}

func newIRCHandler() (*ircHandler, error) {
	h := new(ircHandler)

	username, err := h.fetchTwitchUsername()
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
		i.ExecuteJoins(config.Channels)

	case "NOTICE":
		// NOTICE (Twitch Commands)
		// General notices from the server.
		i.handleTwitchNotice(m)

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

func (ircHandler) fetchTwitchUsername() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), twitchRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		return "", errors.Wrap(err, "assemble user request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", cfg.TwitchClient)
	req.Header.Set("Authorization", "Bearer "+cfg.TwitchToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "requesting user info")
	}
	defer resp.Body.Close()

	var payload struct {
		Data []struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"data"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", errors.Wrap(err, "parse user info")
	}

	if l := len(payload.Data); l != 1 {
		return "", errors.Errorf("unexpected number of users returned: %d", l)
	}

	return payload.Data[0].Login, nil
}

func (i ircHandler) handlePermit(m *irc.Message) {
	badges := i.ParseBadgeLevels(m)
	if !badges.Has(badgeBroadcaster) && (!config.PermitAllowModerator || !badges.Has(badgeModerator)) {
		// Neither broadcaster nor moderator or moderator not permitted
		return
	}

	msgParts := strings.Split(m.Trailing(), " ")
	if len(msgParts) != 2 {
		return
	}

	username := msgParts[1]

	log.WithField("user", username).Debug("Added permit")
	timerStore.Add(timerStore.NormalizeUsername(username))

	go handleMessage(i.c, m, eventTypePermit)
}

func (i ircHandler) handleTwitchNotice(m *irc.Message) {
	log.WithFields(log.Fields{
		"tags":     m.Tags,
		"trailing": m.Trailing,
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
		"name":     m.Name,
		"user":     m.User,
		"tags":     m.Tags,
		"trailing": m.Trailing(),
	}).Trace("Received privmsg")

	if strings.HasPrefix(m.Trailing(), "!permit") {
		i.handlePermit(m)
		return
	}

	go handleMessage(i.c, m, nil)
}

func (i ircHandler) handleTwitchUsernotice(m *irc.Message) {
	log.WithFields(log.Fields{
		"tags":     m.Tags,
		"trailing": m.Trailing,
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

	}
}

func (ircHandler) ParseBadgeLevels(m *irc.Message) badgeCollection {
	out := badgeCollection{}

	badgeString, ok := m.GetTag("badges")
	if !ok {
		return out
	}

	badges := strings.Split(badgeString, ",")
	for _, b := range badges {
		badgeParts := strings.Split(b, "/")
		if len(badgeParts) != 2 {
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

	return out
}
