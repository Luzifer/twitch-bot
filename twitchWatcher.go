package main

import (
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/internal/service/access"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

type (
	topicRegistration struct {
		Topic          string
		Condition      twitch.EventSubCondition
		RequiredScopes []string
		AnyScope       bool
		Hook           func(json.RawMessage) error
		Version        string
	}

	twitchChannelState struct {
		Category string
		IsLive   bool
		Title    string

		isInitialized bool
		esc           *twitch.EventSubSocketClient
	}

	twitchWatcher struct {
		ChannelStatus map[string]*twitchChannelState

		lock sync.RWMutex
	}
)

func (t *twitchChannelState) CloseESC() {
	t.esc.Close()
	t.esc = nil
}

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
		if t.ChannelStatus[c].esc != nil {
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

	if t.ChannelStatus[channel].esc != nil {
		t.ChannelStatus[channel].esc.Close()
	}

	delete(t.ChannelStatus, channel)
	return nil
}

func (t *twitchWatcher) getTopicRegistrations(userID string) []topicRegistration {
	return []topicRegistration{
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
			Version:        twitch.EventSubTopicVersion2,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID, ModeratorUserID: userID},
			RequiredScopes: []string{twitch.ScopeModeratorReadFollowers},
			Hook:           t.handleEventSubChannelFollow,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelRaid,
			Condition:      twitch.EventSubCondition{FromBroadcasterUserID: userID},
			RequiredScopes: nil,
			Hook:           t.handleEventSubChannelOutboundRaid,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelPointCustomRewardRedemptionAdd,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeChannelReadRedemptions, twitch.ScopeChannelManageRedemptions},
			AnyScope:       true,
			Hook:           t.handleEventSubChannelPointCustomRewardRedemptionAdd,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelShoutoutCreate,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID, ModeratorUserID: userID},
			RequiredScopes: []string{twitch.ScopeModeratorManageShoutouts, twitch.ScopeModeratorReadShoutouts},
			AnyScope:       true,
			Hook:           t.handleEventSubShoutoutCreated,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelShoutoutReceive,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID, ModeratorUserID: userID},
			RequiredScopes: []string{twitch.ScopeModeratorManageShoutouts, twitch.ScopeModeratorReadShoutouts},
			AnyScope:       true,
			Hook:           t.handleEventSubShoutoutReceived,
		},
	}
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

func (t *twitchWatcher) handleEventSubChannelOutboundRaid(m json.RawMessage) error {
	var payload twitch.EventSubEventRaid
	if err := json.Unmarshal(m, &payload); err != nil {
		return errors.Wrap(err, "unmarshalling event")
	}

	fields := plugins.FieldCollectionFromData(map[string]interface{}{
		"channel": "#" + payload.FromBroadcasterUserLogin,
		"to_id":   payload.ToBroadcasterUserID,
		"to":      payload.ToBroadcasterUserLogin,
		"viewers": payload.Viewers,
	})

	log.WithFields(log.Fields(fields.Data())).Info("Outbound raid detected")
	go handleMessage(ircHdl.Client(), nil, eventTypeOutboundRaid, fields)

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

func (t *twitchWatcher) handleEventSubShoutoutCreated(m json.RawMessage) error {
	var payload twitch.EventSubEventShoutoutCreated
	if err := json.Unmarshal(m, &payload); err != nil {
		return errors.Wrap(err, "unmarshalling event")
	}

	fields := plugins.FieldCollectionFromData(map[string]any{
		"channel": "#" + payload.BroadcasterUserLogin,
		"to_id":   payload.ToBroadcasterUserID,
		"to":      payload.ToBroadcasterUserLogin,
		"viewers": payload.ViewerCount,
	})

	log.WithFields(log.Fields(fields.Data())).Info("Shoutout created")
	go handleMessage(ircHdl.Client(), nil, eventTypeShoutoutCreated, fields)

	return nil
}

func (t *twitchWatcher) handleEventSubShoutoutReceived(m json.RawMessage) error {
	var payload twitch.EventSubEventShoutoutReceived
	if err := json.Unmarshal(m, &payload); err != nil {
		return errors.Wrap(err, "unmarshalling event")
	}

	fields := plugins.FieldCollectionFromData(map[string]any{
		"channel": "#" + payload.BroadcasterUserLogin,
		"from_id": payload.FromBroadcasterUserID,
		"from":    payload.FromBroadcasterUserLogin,
		"viewers": payload.ViewerCount,
	})

	log.WithFields(log.Fields(fields.Data())).Info("Shoutout received")
	go handleMessage(ircHdl.Client(), nil, eventTypeShoutoutReceived, fields)

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

	if storedStatus.esc != nil {
		// Do not register twice
		return nil
	}

	if storedStatus.esc, err = t.registerEventSubCallbacks(channel); err != nil {
		return errors.Wrap(err, "registering eventsub callbacks")
	}

	if storedStatus.esc != nil {
		log.WithField("channel", channel).Info("watching for eventsub events")
		go func(storedStatus *twitchChannelState) {
			if err := storedStatus.esc.Run(); err != nil {
				log.WithField("channel", channel).WithError(err).Error("eventsub client caused error")
			}
			storedStatus.CloseESC()
		}(storedStatus)
	}

	return nil
}

func (t *twitchWatcher) registerEventSubCallbacks(channel string) (*twitch.EventSubSocketClient, error) {
	tc, err := accessService.GetTwitchClientForChannel(channel, access.ClientConfig{
		TwitchClient:       cfg.TwitchClient,
		TwitchClientSecret: cfg.TwitchClientSecret,
	})
	if err != nil {
		if errors.Is(err, access.ErrChannelNotAuthorized) {
			return nil, nil
		}

		return nil, errors.Wrap(err, "getting twitch client for channel")
	}

	userID, err := twitchClient.GetIDForUsername(channel)
	if err != nil {
		return nil, errors.Wrap(err, "resolving channel to user-id")
	}

	var (
		topicRegistrations = t.getTopicRegistrations(userID)
		topicOpts          []twitch.EventSubSocketClientOpt
	)

	for _, tr := range topicRegistrations {
		logger := log.WithFields(log.Fields{
			"any":     tr.AnyScope,
			"channel": channel,
			"scopes":  tr.RequiredScopes,
			"topic":   tr.Topic,
		})

		if len(tr.RequiredScopes) > 0 {
			fn := accessService.HasPermissionsForChannel
			if tr.AnyScope {
				fn = accessService.HasAnyPermissionForChannel
			}

			hasScopes, err := fn(channel, tr.RequiredScopes...)
			if err != nil {
				return nil, errors.Wrap(err, "checking granted scopes")
			}

			if !hasScopes {
				logger.Debug("Missing scopes for eventsub topic")
				continue
			}
		}

		topicOpts = append(topicOpts, twitch.WithSubscription(tr.Topic, tr.Version, tr.Condition, tr.Hook))
	}

	esClient, err := twitch.NewEventSubSocketClient(append(
		topicOpts,
		twitch.WithLogger(log.WithField("channel", channel)),
		twitch.WithTwitchClient(tc),
	)...)
	if err != nil {
		return nil, errors.Wrap(err, "getting eventsub client for channel")
	}

	return esClient, nil
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
