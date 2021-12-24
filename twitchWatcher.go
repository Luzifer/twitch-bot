package main

import (
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/Luzifer/twitch-bot/twitch"
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

func (t *twitchWatcher) AddChannel(channel string) error {
	t.lock.RLock()
	_, ok := t.ChannelStatus[channel]
	t.lock.RUnlock()

	if ok {
		return nil
	}

	return t.updateChannelFromAPI(channel, false)
}

func (t *twitchWatcher) Check() {
	var channels []string
	t.lock.RLock()
	for c := range t.ChannelStatus {
		if t.ChannelStatus[c].unregisterFunc != nil {
			continue
		}

		channels = append(channels, c)
	}
	t.lock.RUnlock()

	for _, ch := range channels {
		if err := t.updateChannelFromAPI(ch, true); err != nil {
			log.WithError(err).WithField("channel", ch).Error("Unable to update channel status")
		}
	}
}

func (t *twitchWatcher) RemoveChannel(channel string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if f := t.ChannelStatus[channel].unregisterFunc; f != nil {
		f()
	}

	delete(t.ChannelStatus, channel)
	return nil
}

func (t *twitchWatcher) updateChannelFromAPI(channel string, sendUpdate bool) error {
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

	t.lock.Lock()
	defer t.lock.Unlock()

	if t.ChannelStatus[channel] != nil && t.ChannelStatus[channel].Equals(status) {
		return nil
	}

	if sendUpdate && t.ChannelStatus[channel] != nil {
		t.triggerUpdate(channel, &status.Title, &status.Category, &status.IsLive)
		return nil
	}

	if status.unregisterFunc, err = t.registerEventSubCallbacks(channel); err != nil {
		return errors.Wrap(err, "registering eventsub callbacks")
	}

	t.ChannelStatus[channel] = &status
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

	unsubFollow, err := twitchEventSubClient.RegisterEventSubHooks(
		twitch.EventSubEventTypeChannelFollow,
		twitch.EventSubCondition{BroadcasterUserID: userID},
		func(m json.RawMessage) error {
			var payload twitch.EventSubEventFollow
			if err := json.Unmarshal(m, &payload); err != nil {
				return errors.Wrap(err, "unmarshalling event")
			}

			fields := plugins.FieldCollectionFromData(map[string]interface{}{
				"channel":     channel,
				"followed_at": payload.FollowedAt,
				"user_id":     payload.UserID,
				"user":        payload.UserLogin,
			})

			log.WithFields(log.Fields(fields.Data())).Info("User followed")
			go handleMessage(ircHdl.Client(), nil, eventTypeFollow, fields)

			return nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "registering channel-follow eventsub")
	}

	return func() {
		unsubCU()
		unsubSOff()
		unsubSOn()
		unsubFollow()
	}, nil
}

func (t *twitchWatcher) triggerUpdate(channel string, title, category *string, online *bool) {
	if category != nil && t.ChannelStatus[channel].Category != *category {
		t.ChannelStatus[channel].Category = *category
		log.WithFields(log.Fields{
			"channel":  channel,
			"category": *category,
		}).Debug("Twitch metadata changed")
		go handleMessage(ircHdl.Client(), nil, eventTypeTwitchCategoryUpdate, plugins.FieldCollectionFromData(map[string]interface{}{
			"channel":  channel,
			"category": *category,
		}))
	}

	if title != nil && t.ChannelStatus[channel].Title != *title {
		t.ChannelStatus[channel].Title = *title
		log.WithFields(log.Fields{
			"channel": channel,
			"title":   *title,
		}).Debug("Twitch metadata changed")
		go handleMessage(ircHdl.Client(), nil, eventTypeTwitchTitleUpdate, plugins.FieldCollectionFromData(map[string]interface{}{
			"channel": channel,
			"title":   *title,
		}))
	}

	if online != nil && t.ChannelStatus[channel].IsLive != *online {
		t.ChannelStatus[channel].IsLive = *online
		log.WithFields(log.Fields{
			"channel": channel,
			"isLive":  *online,
		}).Debug("Twitch metadata changed")

		evt := eventTypeTwitchStreamOnline
		if !*online {
			evt = eventTypeTwitchStreamOffline
		}

		go handleMessage(ircHdl.Client(), nil, evt, plugins.FieldCollectionFromData(map[string]interface{}{
			"channel": channel,
		}))
	}
}
