package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/v2/fieldcollection"
	"github.com/Luzifer/go_helpers/v2/str"
)

var cronParser = cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

type autoMessage struct {
	UUID string `hash:"-" json:"uuid,omitempty" yaml:"uuid,omitempty"`

	Channel   string `json:"channel,omitempty" yaml:"channel,omitempty"`
	Message   string `json:"message,omitempty" yaml:"message,omitempty"`
	UseAction bool   `json:"use_action,omitempty" yaml:"use_action,omitempty"`

	Disable           bool    `json:"disable,omitempty" yaml:"disable,omitempty"`
	DisableOnTemplate *string `json:"disable_on_template,omitempty" yaml:"disable_on_template,omitempty"`

	Cron            string `json:"cron,omitempty" yaml:"cron,omitempty"`
	MessageInterval int64  `json:"message_interval,omitempty" yaml:"message_interval,omitempty"`
	OnlyOnLive      bool   `json:"only_on_live,omitempty" yaml:"only_on_live,omitempty"`

	disabled              bool
	lastMessageSent       time.Time
	linesSinceLastMessage int64

	lock sync.RWMutex
}

func (a *autoMessage) CanSend() bool {
	a.lock.RLock()
	defer a.lock.RUnlock()

	if a.disabled || !a.IsValid() {
		return false
	}

	switch {
	case !str.StringInSlice(a.Channel, config.Channels):
		// Not an observed channel, auto-message is not valid
		return false

	case a.MessageInterval > a.linesSinceLastMessage:
		// Not enough chatted lines
		return false

	case a.Cron != "":
		sched, _ := cronParser.Parse(a.Cron)
		nextExecute := sched.Next(a.lastMessageSent)
		if nextExecute.After(time.Now()) {
			// Cron timer is not yet expired
			return false
		}
		log.WithFields(log.Fields{
			"lastMessage":   a.lastMessageSent,
			"nextExecution": nextExecute,
			"now":           time.Now(),
		}).Trace("Auto-Message was allowed through cron")
	}

	if a.OnlyOnLive {
		streamLive, err := twitchClient.HasLiveStream(context.Background(), strings.TrimLeft(a.Channel, "#"))
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

	if a.Disable {
		log.Trace("Auto-Message disabled by flag")
		// Reset the timer for this execution not to spam every second
		a.lastMessageSent = time.Now()
		return false
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
	if a.UUID != "" {
		return a.UUID
	}

	h, err := hashstructure.Hash(a, hashstructure.FormatV2, nil)
	if err != nil {
		panic(errors.Wrap(err, "hashing automessage"))
	}
	return fmt.Sprintf("hashstructure:%x", h)
}

func (a *autoMessage) IsValid() bool {
	if a.Cron != "" {
		if _, err := cronParser.Parse(a.Cron); err != nil {
			return false
		}
	}

	if a.MessageInterval == 0 && a.Cron == "" {
		return false
	}

	return true
}

func (a *autoMessage) Send(_ *irc.Client) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	msg, err := formatMessage(a.Message, nil, nil, nil)
	if err != nil {
		return errors.Wrap(err, "preparing message")
	}

	if a.UseAction {
		msg = fmt.Sprintf("\001ACTION %s\001", msg)
	}

	if err := sendMessage(&irc.Message{
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
	if a.DisableOnTemplate == nil || *a.DisableOnTemplate == "" {
		// No match criteria set, does not speak against matching
		return true
	}

	fields := fieldcollection.NewFieldCollection()
	fields.Set("channel", a.Channel)

	res, err := formatMessage(*a.DisableOnTemplate, nil, nil, fields)
	if err != nil {
		log.WithError(err).Error("Error in auto-message disable template")
		// Caused an error, forbid execution
		return false
	}

	return res != "true"
}
