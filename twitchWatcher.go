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

		isInitialized  bool
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

func (t *twitchChannelState) Update(c twitchChannelState) {
	t.Category = c.Category
	t.IsLive = c.IsLive
	t.Title = c.Title
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

	// Initialize for check loop
	t.lock.Lock()
	t.ChannelStatus[channel] = &twitchChannelState{}
	t.lock.Unlock()

	return t.updateChannelFromAPI(channel)
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
		if err := t.updateChannelFromAPI(ch); err != nil {
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

func (t *twitchWatcher) handleEventSubChannelFollow(m json.RawMessage) error {
	var payload twitch.EventSubEventFollow
	if err := json.Unmarshal(m, &payload); err != nil {
		return errors.Wrap(err, "unmarshalling event")
	}

	fields := plugins.FieldCollectionFromData(map[string]interface{}{
		"channel":     "#" + payload.BroadcasterUserLogin,
		"followed_at": payload.FollowedAt,
		"user_id":     payload.UserID,
		"user":        payload.UserLogin,
	})

	log.WithFields(log.Fields(fields.Data())).Info("User followed")
	go handleMessage(ircHdl.Client(), nil, eventTypeFollow, fields)

	return nil
}

func (t *twitchWatcher) handleEventSubChannelPointCustomRewardRedemptionAdd(m json.RawMessage) error {
	var payload twitch.EventSubEventChannelPointCustomRewardRedemptionAdd
	if err := json.Unmarshal(m, &payload); err != nil {
		return errors.Wrap(err, "unmarshalling event")
	}

	fields := plugins.FieldCollectionFromData(map[string]interface{}{
		"channel":      "#" + payload.BroadcasterUserLogin,
		"reward_cost":  payload.Reward.Cost,
		"reward_id":    payload.Reward.ID,
		"reward_title": payload.Reward.Title,
		"status":       payload.Status,
		"user_id":      payload.UserID,
		"user_input":   payload.UserInput,
		"user":         payload.UserLogin,
	})

	log.WithFields(log.Fields(fields.Data())).Info("ChannelPoint reward was redeemed")
	go handleMessage(ircHdl.Client(), nil, eventTypeChannelPointRedeem, fields)

	return nil
}

func (t *twitchWatcher) handleEventSubChannelUpdate(m json.RawMessage) error {
	var payload twitch.EventSubEventChannelUpdate
	if err := json.Unmarshal(m, &payload); err != nil {
		return errors.Wrap(err, "unmarshalling event")
	}

	t.triggerUpdate(payload.BroadcasterUserLogin, &payload.Title, &payload.CategoryName, nil)

	return nil
}

func (t *twitchWatcher) handleEventSubStreamOnOff(isOnline bool) func(json.RawMessage) error {
	return func(m json.RawMessage) error {
		var payload twitch.EventSubEventFollow
		if err := json.Unmarshal(m, &payload); err != nil {
			return errors.Wrap(err, "unmarshalling event")
		}

		t.triggerUpdate(payload.BroadcasterUserLogin, nil, nil, &isOnline)
		return nil
	}
}

func (t *twitchWatcher) handleEventUserAuthRevoke(m json.RawMessage) error {
	var payload twitch.EventSubEventUserAuthorizationRevoke
	if err := json.Unmarshal(m, &payload); err != nil {
		return errors.Wrap(err, "unmarshalling event")
	}

	if payload.ClientID != cfg.TwitchClient {
		// We got an revoke for a different ID: Shouldn't happen but whatever.
		return nil
	}

	return errors.Wrap(
		store.DeleteGrantedScopes(payload.UserLogin),
		"deleting granted scopes",
	)
}

func (t *twitchWatcher) updateChannelFromAPI(channel string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	var (
		err          error
		status       twitchChannelState
		storedStatus = t.ChannelStatus[channel]
	)

	status.IsLive, err = twitchClient.HasLiveStream(channel)
	if err != nil {
		return errors.Wrap(err, "getting live status")
	}

	status.Category, status.Title, err = twitchClient.GetRecentStreamInfo(channel)
	if err != nil {
		return errors.Wrap(err, "getting stream info")
	}

	if storedStatus == nil {
		storedStatus = &twitchChannelState{}
		t.ChannelStatus[channel] = storedStatus
	}

	if storedStatus.isInitialized && !storedStatus.Equals(status) {
		// Send updates only when we do have an update
		t.triggerUpdate(channel, &status.Title, &status.Category, &status.IsLive)
	}

	storedStatus.Update(status)
	storedStatus.isInitialized = true

	if storedStatus.unregisterFunc != nil {
		// Do not register twice
		return nil
	}

	if storedStatus.unregisterFunc, err = t.registerEventSubCallbacks(channel); err != nil {
		return errors.Wrap(err, "registering eventsub callbacks")
	}

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

	var (
		topicRegistrations = []struct {
			Topic          string
			Condition      twitch.EventSubCondition
			RequiredScopes []string
			AnyScope       bool
			Hook           func(json.RawMessage) error
		}{
			{
				Topic:          twitch.EventSubEventTypeChannelUpdate,
				Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
				RequiredScopes: nil,
				Hook:           t.handleEventSubChannelUpdate,
			},
			{
				Topic:          twitch.EventSubEventTypeStreamOffline,
				Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
				RequiredScopes: nil,
				Hook:           t.handleEventSubStreamOnOff(false),
			},
			{
				Topic:          twitch.EventSubEventTypeStreamOnline,
				Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
				RequiredScopes: nil,
				Hook:           t.handleEventSubStreamOnOff(true),
			},
			{
				Topic:          twitch.EventSubEventTypeChannelFollow,
				Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
				RequiredScopes: nil,
				Hook:           t.handleEventSubChannelFollow,
			},
			{
				Topic:          twitch.EventSubEventTypeChannelPointCustomRewardRedemptionAdd,
				Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
				RequiredScopes: []string{twitch.ScopeChannelReadRedemptions, twitch.ScopeChannelManageRedemptions},
				AnyScope:       true,
				Hook:           t.handleEventSubChannelPointCustomRewardRedemptionAdd,
			},
		}
		unsubHandlers []func()
	)

	for _, tr := range topicRegistrations {
		logger := log.WithFields(log.Fields{
			"any":     tr.AnyScope,
			"channel": channel,
			"scopes":  tr.RequiredScopes,
			"topic":   tr.Topic,
		})

		if len(tr.RequiredScopes) > 0 {
			fn := store.UserHasGrantedScopes
			if tr.AnyScope {
				fn = store.UserHasGrantedAnyScope
			}

			if !fn(channel, tr.RequiredScopes...) {
				logger.Debug("Missing scopes for eventsub topic")
				continue
			}
		}

		uf, err := twitchEventSubClient.RegisterEventSubHooks(tr.Topic, tr.Condition, tr.Hook)
		if err != nil {
			logger.WithError(err).Error("Unable to register topic")

			for _, f := range unsubHandlers {
				// Error will cause unsub handlers not to be stored, therefore we unsub them now
				f()
			}

			return nil, errors.Wrap(err, "registering topic")
		}

		unsubHandlers = append(unsubHandlers, uf)
	}

	return func() {
		for _, f := range unsubHandlers {
			f()
		}
	}, nil
}

func (t *twitchWatcher) registerGlobalHooks() error {
	_, err := twitchEventSubClient.RegisterEventSubHooks(
		twitch.EventSubEventTypeUserAuthorizationRevoke,
		twitch.EventSubCondition{ClientID: cfg.TwitchClient},
		t.handleEventUserAuthRevoke,
	)

	return errors.Wrap(err, "registering user auth hook")
}

func (t *twitchWatcher) triggerUpdate(channel string, title, category *string, online *bool) {
	if category != nil && t.ChannelStatus[channel].Category != *category {
		t.ChannelStatus[channel].Category = *category
		log.WithFields(log.Fields{
			"channel":  channel,
			"category": *category,
		}).Debug("Twitch metadata changed")
		go handleMessage(ircHdl.Client(), nil, eventTypeTwitchCategoryUpdate, plugins.FieldCollectionFromData(map[string]interface{}{
			"channel":  "#" + channel,
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
			"channel": "#" + channel,
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
			"channel": "#" + channel,
		}))
	}
}
