package main

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/go_helpers/v2/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/internal/service/access"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
)

type (
	topicRegistration struct {
		Topic          string
		Condition      twitch.EventSubCondition
		RequiredScopes []string
		AnyScope       bool
		Hook           func(json.RawMessage) error
		Version        string
		Optional       bool
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

//nolint:funlen // Just a collection of topics
func (t *twitchWatcher) getTopicRegistrations(userID string) []topicRegistration {
	return []topicRegistration{
		{
			Topic:          twitch.EventSubEventTypeChannelAdBreakBegin,
			Version:        twitch.EventSubTopicVersion1,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeChannelReadAds},
			Hook:           t.handleEventSubChannelAdBreakBegin,
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelFollow,
			Version:        twitch.EventSubTopicVersion2,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID, ModeratorUserID: userID},
			RequiredScopes: []string{twitch.ScopeModeratorReadFollowers},
			Hook:           t.handleEventSubChannelFollow,
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelHypetrainBegin,
			Version:        twitch.EventSubTopicVersion2,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeChannelReadHypetrain},
			Hook:           t.handleEventSubHypetrainEvent(eventTypeHypetrainBegin),
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelHypetrainEnd,
			Version:        twitch.EventSubTopicVersion2,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeChannelReadHypetrain},
			Hook:           t.handleEventSubHypetrainEvent(eventTypeHypetrainEnd),
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelHypetrainProgress,
			Version:        twitch.EventSubTopicVersion2,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeChannelReadHypetrain},
			Hook:           t.handleEventSubHypetrainEvent(eventTypeHypetrainProgress),
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelPointCustomRewardRedemptionAdd,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeChannelReadRedemptions, twitch.ScopeChannelManageRedemptions},
			AnyScope:       true,
			Hook:           t.handleEventSubChannelPointCustomRewardRedemptionAdd,
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelPollBegin,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeChannelReadPolls, twitch.ScopeChannelManagePolls},
			AnyScope:       true,
			Hook:           t.handleEventSubChannelPollChange(eventTypePollBegin),
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelPollEnd,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeChannelReadPolls, twitch.ScopeChannelManagePolls},
			AnyScope:       true,
			Hook:           t.handleEventSubChannelPollChange(eventTypePollEnd),
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelPollProgress,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeChannelReadPolls, twitch.ScopeChannelManagePolls},
			AnyScope:       true,
			Hook:           t.handleEventSubChannelPollChange(eventTypePollProgress),
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelRaid,
			Condition:      twitch.EventSubCondition{FromBroadcasterUserID: userID},
			RequiredScopes: nil,
			Hook:           t.handleEventSubChannelOutboundRaid,
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelShoutoutCreate,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID, ModeratorUserID: userID},
			RequiredScopes: []string{twitch.ScopeModeratorManageShoutouts, twitch.ScopeModeratorReadShoutouts},
			AnyScope:       true,
			Hook:           t.handleEventSubShoutoutCreated,
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelShoutoutReceive,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID, ModeratorUserID: userID},
			RequiredScopes: []string{twitch.ScopeModeratorManageShoutouts, twitch.ScopeModeratorReadShoutouts},
			AnyScope:       true,
			Hook:           t.handleEventSubShoutoutReceived,
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelUpdate,
			Version:        twitch.EventSubTopicVersion2,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: nil,
			Hook:           t.handleEventSubChannelUpdate,
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeStreamOffline,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: nil,
			Hook:           t.handleEventSubStreamOnOff(false),
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeStreamOnline,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: nil,
			Hook:           t.handleEventSubStreamOnOff(true),
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelSuspiciousUserMessage,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID, ModeratorUserID: userID},
			RequiredScopes: []string{twitch.ScopeModeratorReadSuspiciousUsers},
			Hook:           t.handleEventSubSusUserMessage,
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelSuspiciousUserUpdate,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID, ModeratorUserID: userID},
			RequiredScopes: []string{twitch.ScopeModeratorReadSuspiciousUsers},
			Hook:           t.handleEventSubSusUserUpdate,
			Optional:       true,
		},
	}
}

func (*twitchWatcher) handleEventSubChannelAdBreakBegin(m json.RawMessage) error {
	var payload twitch.EventSubEventAdBreakBegin
	if err := json.Unmarshal(m, &payload); err != nil {
		return errors.Wrap(err, "unmarshalling event")
	}

	fields := fieldcollection.FieldCollectionFromData(map[string]any{
		"channel":      "#" + payload.BroadcasterUserLogin,
		"duration":     payload.Duration,
		"is_automatic": payload.IsAutomatic,
		"started_at":   payload.StartedAt,
	})

	log.WithFields(log.Fields(fields.Data())).Info("Ad-Break started")
	go handleMessage(ircHdl.Client(), nil, eventTypeAdBreakBegin, fields)

	return nil
}

func (*twitchWatcher) handleEventSubChannelFollow(m json.RawMessage) error {
	var payload twitch.EventSubEventFollow
	if err := json.Unmarshal(m, &payload); err != nil {
		return errors.Wrap(err, "unmarshalling event")
	}

	fields := fieldcollection.FieldCollectionFromData(map[string]interface{}{
		"channel":     "#" + payload.BroadcasterUserLogin,
		"followed_at": payload.FollowedAt,
		"user_id":     payload.UserID,
		"user":        payload.UserLogin,
	})

	log.WithFields(log.Fields(fields.Data())).Info("User followed")
	go handleMessage(ircHdl.Client(), nil, eventTypeFollow, fields)

	return nil
}

func (*twitchWatcher) handleEventSubChannelPointCustomRewardRedemptionAdd(m json.RawMessage) error {
	var payload twitch.EventSubEventChannelPointCustomRewardRedemptionAdd
	if err := json.Unmarshal(m, &payload); err != nil {
		return errors.Wrap(err, "unmarshalling event")
	}

	fields := fieldcollection.FieldCollectionFromData(map[string]interface{}{
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

func (*twitchWatcher) handleEventSubChannelOutboundRaid(m json.RawMessage) error {
	var payload twitch.EventSubEventRaid
	if err := json.Unmarshal(m, &payload); err != nil {
		return errors.Wrap(err, "unmarshalling event")
	}

	fields := fieldcollection.FieldCollectionFromData(map[string]interface{}{
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

func (*twitchWatcher) handleEventSubChannelPollChange(event *string) func(json.RawMessage) error {
	return func(m json.RawMessage) error {
		var payload twitch.EventSubEventPoll
		if err := json.Unmarshal(m, &payload); err != nil {
			return errors.Wrap(err, "unmarshalling event")
		}

		fields := fieldcollection.FieldCollectionFromData(map[string]any{
			"channel":               "#" + payload.BroadcasterUserLogin,
			"hasChannelPointVoting": payload.ChannelPointsVoting.IsEnabled,
			"title":                 payload.Title,
		})

		logger := log.WithFields(log.Fields(fields.Data()))

		switch event {
		case eventTypePollBegin:
			logger.Info("Poll started")

		case eventTypePollEnd:
			fields.Set("status", payload.Status)
			logger.WithField("status", payload.Status).Info("Poll ended")

		case eventTypePollProgress:
			// Lets not spam the info-level-log with every single vote but
			// provide them for bots with debug-level-logging
			logger.Debug("Poll changed")
		}

		// Set after logging not to spam logs with full payload
		fields.Set("poll", payload)

		go handleMessage(ircHdl.Client(), nil, event, fields)
		return nil
	}
}

func (*twitchWatcher) handleEventSubHypetrainEvent(eventType *string) func(json.RawMessage) error {
	return func(m json.RawMessage) error {
		var payload twitch.EventSubEventHypetrain
		if err := json.Unmarshal(m, &payload); err != nil {
			return errors.Wrap(err, "unmarshalling event")
		}

		fields := fieldcollection.FieldCollectionFromData(map[string]any{
			"channel": "#" + payload.BroadcasterUserLogin,
			"level":   payload.Level,
		})

		if payload.Goal > 0 {
			fields.Set("levelProgress", float64(payload.Progress)/float64(payload.Goal))
		}

		log.WithFields(log.Fields(fields.Data())).Info("Hypetrain event")

		fields.Set("event", payload)
		go handleMessage(ircHdl.Client(), nil, eventType, fields)

		return nil
	}
}

func (*twitchWatcher) handleEventSubShoutoutCreated(m json.RawMessage) error {
	var payload twitch.EventSubEventShoutoutCreated
	if err := json.Unmarshal(m, &payload); err != nil {
		return errors.Wrap(err, "unmarshalling event")
	}

	fields := fieldcollection.FieldCollectionFromData(map[string]any{
		"channel": "#" + payload.BroadcasterUserLogin,
		"to_id":   payload.ToBroadcasterUserID,
		"to":      payload.ToBroadcasterUserLogin,
		"viewers": payload.ViewerCount,
	})

	log.WithFields(log.Fields(fields.Data())).Info("Shoutout created")
	go handleMessage(ircHdl.Client(), nil, eventTypeShoutoutCreated, fields)

	return nil
}

func (*twitchWatcher) handleEventSubShoutoutReceived(m json.RawMessage) error {
	var payload twitch.EventSubEventShoutoutReceived
	if err := json.Unmarshal(m, &payload); err != nil {
		return errors.Wrap(err, "unmarshalling event")
	}

	fields := fieldcollection.FieldCollectionFromData(map[string]any{
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

func (*twitchWatcher) handleEventSubSusUserMessage(m json.RawMessage) (err error) {
	var payload twitch.EventSubEventSuspiciousUserMessage
	if err := json.Unmarshal(m, &payload); err != nil {
		return errors.Wrap(err, "unmarshalling event")
	}

	fields := fieldcollection.FieldCollectionFromData(map[string]any{
		"ban_evasion":         payload.BanEvasionEvaluation,
		"channel":             "#" + payload.BroadcasterUserLogin,
		"message":             payload.Message.Text,
		"shared_ban_channels": payload.SharedBanChannelIDs,
		"status":              payload.LowTrustStatus,
		"user_id":             payload.UserID,
		"user_type":           payload.Types,
		"username":            payload.UserLogin,
	})

	log.WithFields(log.Fields(fields.Data())).Info("restricted user message")
	go handleMessage(ircHdl.Client(), nil, eventTypeSusUserMessage, fields)

	return nil
}

func (*twitchWatcher) handleEventSubSusUserUpdate(m json.RawMessage) (err error) {
	var payload twitch.EventSubEventSuspiciousUserUpdated
	if err := json.Unmarshal(m, &payload); err != nil {
		return errors.Wrap(err, "unmarshalling event")
	}

	fields := fieldcollection.FieldCollectionFromData(map[string]any{
		"channel":   "#" + payload.BroadcasterUserLogin,
		"moderator": payload.ModeratorUserLogin,
		"status":    payload.LowTrustStatus,
		"user_id":   payload.UserID,
		"username":  payload.UserLogin,
	})

	log.WithFields(log.Fields(fields.Data())).Info("user restriction updated")
	go handleMessage(ircHdl.Client(), nil, eventTypeSusUserUpdate, fields)

	return nil
}

func (t *twitchWatcher) updateChannelFromAPI(channel string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	var (
		err          error
		status       twitchChannelState
		storedStatus = t.ChannelStatus[channel]
	)

	status.IsLive, err = twitchClient.HasLiveStream(context.Background(), channel)
	if err != nil {
		return errors.Wrap(err, "getting live status")
	}

	status.Category, status.Title, err = twitchClient.GetRecentStreamInfo(context.Background(), channel)
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
				log.WithField("channel", channel).WithError(helpers.CleanNetworkAddressFromError(err)).Error("eventsub client caused error")
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
			return nil, nil //nolint:nilnil // This is fine - not authorized cannot register callbacks
		}

		return nil, errors.Wrap(err, "getting twitch client for channel")
	}

	userID, err := twitchClient.GetIDForUsername(context.Background(), channel)
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

		var opt twitch.EventSubSocketClientOpt
		if tr.Optional {
			opt = twitch.WithRetryBackgroundSubscribe(tr.Topic, tr.Version, tr.Condition, tr.Hook)
		} else {
			opt = twitch.WithMustSubscribe(tr.Topic, tr.Version, tr.Condition, tr.Hook)
		}

		topicOpts = append(topicOpts, opt)
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
		}).Info("Category updated")
		go handleMessage(ircHdl.Client(), nil, eventTypeTwitchCategoryUpdate, fieldcollection.FieldCollectionFromData(map[string]interface{}{
			"channel":  "#" + channel,
			"category": *category,
		}))
	}

	if title != nil && t.ChannelStatus[channel].Title != *title {
		t.ChannelStatus[channel].Title = *title
		log.WithFields(log.Fields{
			"channel": channel,
			"title":   *title,
		}).Info("Title updated")
		go handleMessage(ircHdl.Client(), nil, eventTypeTwitchTitleUpdate, fieldcollection.FieldCollectionFromData(map[string]interface{}{
			"channel": "#" + channel,
			"title":   *title,
		}))
	}

	if online != nil && t.ChannelStatus[channel].IsLive != *online {
		t.ChannelStatus[channel].IsLive = *online
		log.WithFields(log.Fields{
			"channel": channel,
			"isLive":  *online,
		}).Info("Live-status updated")

		evt := eventTypeTwitchStreamOnline
		if !*online {
			evt = eventTypeTwitchStreamOffline
		}

		go handleMessage(ircHdl.Client(), nil, evt, fieldcollection.FieldCollectionFromData(map[string]interface{}{
			"channel": "#" + channel,
		}))
	}
}
