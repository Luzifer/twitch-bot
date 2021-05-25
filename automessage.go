package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

var cronParser = cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

type autoMessage struct {
	Channel   string `yaml:"channel"`
	Message   string `yaml:"message"`
	UseAction bool   `yaml:"use_action"`

	DisableOnTemplate *string `yaml:"disable_on_template"`

	Cron            string        `yaml:"cron"`
	MessageInterval int64         `yaml:"message_interval"`
	OnlyOnLive      bool          `yaml:"only_on_live"`
	TimeInterval    time.Duration `yaml:"time_interval"`

	disabled              bool
	lastMessageSent       time.Time
	linesSinceLastMessage int64

	lock sync.RWMutex
}

func (a *autoMessage) CanSend() bool {
	if a.disabled || !a.IsValid() {
		return false
	}

	a.lock.RLock()
	defer a.lock.RUnlock()

	switch {
	case a.MessageInterval > a.linesSinceLastMessage:
		// Not enough chatted lines
		return false

	case a.TimeInterval > 0 && a.lastMessageSent.Add(a.TimeInterval).After(time.Now()):
		// Simple timer is not yet expired
		return false

	case a.Cron != "":
		sched, _ := cronParser.Parse(a.Cron)
		if sched.Next(a.lastMessageSent).After(time.Now()) {
			// Cron timer is not yet expired
			return false
		}
	}

	if a.OnlyOnLive {
		streamLive, err := twitch.HasLiveStream(strings.TrimLeft(a.Channel, "#"))
		if err != nil {
			log.WithError(err).Error("Unable to determine channel live status")
			return false
		}
		if !streamLive {
			// Timer is only to be triggered during stream being live,
			// reset the timer in order not to spam all messages on stream-start
			a.lastMessageSent = time.Now()
			return false
		}
	}

	if !a.allowExecuteDisableOnTemplate() {
		log.Trace("Auto-Message disabled by template")
		// Reset the timer for this execution not to spam every second
		a.lastMessageSent = time.Now()
		return false
	}

	return true
}

func (a *autoMessage) CountMessage(channel string) {
	if strings.TrimLeft(channel, "#") != strings.TrimLeft(a.Channel, "#") {
		return
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	a.linesSinceLastMessage++
}

func (a *autoMessage) ID() string {
	sum := sha256.New()

	fmt.Fprintf(sum, "channel:%q", a.Channel)
	fmt.Fprintf(sum, "message:%q", a.Message)
	fmt.Fprintf(sum, "action:%v", a.UseAction)

	return fmt.Sprintf("sha256:%x", sum.Sum(nil))
}

func (a *autoMessage) IsValid() bool {
	if a.Cron != "" {
		if _, err := cronParser.Parse(a.Cron); err != nil {
			return false
		}
	}

	if a.MessageInterval == 0 && a.TimeInterval == 0 && a.Cron == "" {
		return false
	}

	return true
}

func (a *autoMessage) Send(c *irc.Client) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	msg, err := formatMessage(a.Message, nil, nil, nil)
	if err != nil {
		return errors.Wrap(err, "preparing message")
	}

	if a.UseAction {
		msg = fmt.Sprintf("\001ACTION %s\001", msg)
	}

	if err := c.WriteMessage(&irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			fmt.Sprintf("#%s", strings.TrimLeft(a.Channel, "#")),
			msg,
		},
	}); err != nil {
		return errors.Wrap(err, "sending auto-message")
	}

	a.lastMessageSent = time.Now()
	a.linesSinceLastMessage = 0

	return nil
}

func (a *autoMessage) allowExecuteDisableOnTemplate() bool {
	if a.DisableOnTemplate == nil {
		// No match criteria set, does not speak against matching
		return true
	}

	res, err := formatMessage(*a.DisableOnTemplate, nil, nil, map[string]interface{}{
		"channel": a.Channel,
	})
	if err != nil {
		log.WithError(err).Error("Error in auto-message disable template")
		// Caused an error, forbid execution
		return false
	}

	return res != "true"
}
