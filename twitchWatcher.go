package main

import (
	"encoding/json"
	"sync"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/Luzifer/twitch-bot/twitch"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type (
	twitchChannelState struct {
		Category string
		IsLive   bool
		Title    string

		unregisterFunc func()
	}

	twitchWatcher struct {
		ChannelStatus map[string]*twitchChannelState

		lock sync.RWMutex
	}
)

func (t twitchChannelState) Equals(c twitchChannelState) bool {
	return t.Category == c.Category &&
		t.IsLive == c.IsLive &&
		t.Title == c.Title
}

func newTwitchWatcher() *twitchWatcher {
	return &twitchWatcher{
		ChannelStatus: make(map[string]*twitchChannelState),
	}
}

func (r *twitchWatcher) AddChannel(channel string) error {
	r.lock.RLock()
	_, ok := r.ChannelStatus[channel]
	r.lock.RUnlock()

	if ok {
		return nil
	}

	return r.updateChannelFromAPI(channel, false)
}

func (r *twitchWatcher) Check() {
	var channels []string
	r.lock.RLock()
	for c := range r.ChannelStatus {
		if r.ChannelStatus[c].unregisterFunc != nil {
			continue
		}

		channels = append(channels, c)
	}
	r.lock.RUnlock()

	for _, ch := range channels {
		if err := r.updateChannelFromAPI(ch, true); err != nil {
			log.WithError(err).WithField("channel", ch).Error("Unable to update channel status")
		}
	}
}

func (r *twitchWatcher) RemoveChannel(channel string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if f := r.ChannelStatus[channel].unregisterFunc; f != nil {
		f()
	}

	delete(r.ChannelStatus, channel)
	return nil
}

func (r *twitchWatcher) updateChannelFromAPI(channel string, sendUpdate bool) error {
	var (
		err    error
		status twitchChannelState
	)

	status.IsLive, err = twitchClient.HasLiveStream(channel)
	if err != nil {
		return errors.Wrap(err, "getting live status")
	}

	status.Category, status.Title, err = twitchClient.GetRecentStreamInfo(channel)
	if err != nil {
		return errors.Wrap(err, "getting stream info")
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	if r.ChannelStatus[channel] != nil && r.ChannelStatus[channel].Equals(status) {
		return nil
	}

	if sendUpdate && r.ChannelStatus[channel] != nil {
		r.triggerUpdate(channel, &status.Title, &status.Category, &status.IsLive)
		return nil
	}

	if status.unregisterFunc, err = r.registerEventSubCallbacks(channel); err != nil {
		return errors.Wrap(err, "registering eventsub callbacks")
	}

	r.ChannelStatus[channel] = &status
	return nil
}

func (t *twitchWatcher) registerEventSubCallbacks(channel string) (func(), error) {
	if twitchEventSubClient == nil {
		// We don't have eventsub functionality
		return nil, nil
	}

	userID, err := twitchClient.GetIDForUsername(channel)
	if err != nil {
		return nil, errors.Wrap(err, "resolving channel to user-id")
	}

	unsubCU, err := twitchEventSubClient.RegisterEventSubHooks(
		twitch.EventSubEventTypeChannelUpdate,
		twitch.EventSubCondition{BroadcasterUserID: userID},
		func(m json.RawMessage) error {
			var payload twitch.EventSubEventChannelUpdate
			if err := json.Unmarshal(m, &payload); err != nil {
				return errors.Wrap(err, "unmarshalling event")
			}

			t.triggerUpdate(channel, &payload.Title, &payload.CategoryName, nil)

			return nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "registering channel-update eventsub")
	}

	unsubSOff, err := twitchEventSubClient.RegisterEventSubHooks(
		twitch.EventSubEventTypeStreamOffline,
		twitch.EventSubCondition{BroadcasterUserID: userID},
		func(m json.RawMessage) error {
			var payload twitch.EventSubEventStreamOffline
			if err := json.Unmarshal(m, &payload); err != nil {
				return errors.Wrap(err, "unmarshalling event")
			}

			t.triggerUpdate(channel, nil, nil, func(v bool) *bool { return &v }(false))

			return nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "registering channel-update eventsub")
	}

	unsubSOn, err := twitchEventSubClient.RegisterEventSubHooks(
		twitch.EventSubEventTypeStreamOnline,
		twitch.EventSubCondition{BroadcasterUserID: userID},
		func(m json.RawMessage) error {
			var payload twitch.EventSubEventStreamOnline
			if err := json.Unmarshal(m, &payload); err != nil {
				return errors.Wrap(err, "unmarshalling event")
			}

			t.triggerUpdate(channel, nil, nil, func(v bool) *bool { return &v }(true))

			return nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "registering channel-update eventsub")
	}

	return func() {
		unsubCU()
		unsubSOff()
		unsubSOn()
	}, nil
}

func (r *twitchWatcher) triggerUpdate(channel string, title, category *string, online *bool) {
	if category != nil && r.ChannelStatus[channel].Category != *category {
		r.ChannelStatus[channel].Category = *category
		log.WithFields(log.Fields{
			"channel":  channel,
			"category": *category,
		}).Debug("Twitch metadata changed")
		go handleMessage(ircHdl.Client(), nil, eventTypeTwitchCategoryUpdate, plugins.FieldCollection{
			"channel":  channel,
			"category": *category,
		})
	}

	if title != nil && r.ChannelStatus[channel].Title != *title {
		r.ChannelStatus[channel].Title = *title
		log.WithFields(log.Fields{
			"channel": channel,
			"title":   *title,
		}).Debug("Twitch metadata changed")
		go handleMessage(ircHdl.Client(), nil, eventTypeTwitchTitleUpdate, plugins.FieldCollection{
			"channel": channel,
			"title":   *title,
		})
	}

	if online != nil && r.ChannelStatus[channel].IsLive != *online {
		r.ChannelStatus[channel].IsLive = *online
		log.WithFields(log.Fields{
			"channel": channel,
			"isLive":  *online,
		}).Debug("Twitch metadata changed")

		evt := eventTypeTwitchStreamOnline
		if !*online {
			evt = eventTypeTwitchStreamOffline
		}

		go handleMessage(ircHdl.Client(), nil, evt, plugins.FieldCollection{
			"channel": channel,
		})
	}
}
