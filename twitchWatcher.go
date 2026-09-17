package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/internal/service/access"
	"github.com/Luzifer/twitch-bot/v3/pkg/event"
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
			Hook:           t.handleEventSubHypetrainEvent(*event.HypetrainBegin{}.Event()),
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelHypetrainEnd,
			Version:        twitch.EventSubTopicVersion2,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeChannelReadHypetrain},
			Hook:           t.handleEventSubHypetrainEvent(*event.HypetrainEnd{}.Event()),
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelHypetrainProgress,
			Version:        twitch.EventSubTopicVersion2,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeChannelReadHypetrain},
			Hook:           t.handleEventSubHypetrainEvent(*event.HypetrainProgress{}.Event()),
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelModeratorAdd,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeModerationRead},
			Hook:           t.handleEventSubModeratorAdd,
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelModeratorRemove,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeModerationRead},
			Hook:           t.handleEventSubModeratorRemove,
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
			Hook:           t.handleEventSubChannelPollChange(*event.PollBegin{}.Event()),
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelPollEnd,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeChannelReadPolls, twitch.ScopeChannelManagePolls},
			AnyScope:       true,
			Hook:           t.handleEventSubChannelPollChange(*event.PollEnd{}.Event()),
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeChannelPollProgress,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeChannelReadPolls, twitch.ScopeChannelManagePolls},
			AnyScope:       true,
			Hook:           t.handleEventSubChannelPollChange(*event.PollProgress{}.Event()),
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
		{
			Topic:          twitch.EventSubEventTypeVIPAdd,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeChannelReadVIPs, twitch.ScopeChannelManageVIPS},
			AnyScope:       true,
			Hook:           t.handleEventSubVIPAdd,
			Optional:       true,
		},
		{
			Topic:          twitch.EventSubEventTypeVIPRemove,
			Condition:      twitch.EventSubCondition{BroadcasterUserID: userID},
			RequiredScopes: []string{twitch.ScopeChannelReadVIPs, twitch.ScopeChannelManageVIPS},
			AnyScope:       true,
			Hook:           t.handleEventSubVIPRemove,
			Optional:       true,
		},
	}
}

func (*twitchWatcher) handleEventSubChannelAdBreakBegin(m json.RawMessage) error {
	var payload twitch.EventSubEventAdBreakBegin
	if err := json.Unmarshal(m, &payload); err != nil {
		return fmt.Errorf("unmarshalling event: %w", err)
	}

	evt := event.AdBreakBegin{
		Channel:     "#" + payload.BroadcasterUserLogin,
		Duration:    payload.Duration,
		IsAutomatic: payload.IsAutomatic,
		StartedAt:   payload.StartedAt,
	}

	log.WithFields(event.ToLogFields(evt)).Info("Ad-Break started")
	go handleTypedMessage(ircHdl.Client(), nil, evt)

	return nil
}

func (*twitchWatcher) handleEventSubChannelFollow(m json.RawMessage) error {
	var payload twitch.EventSubEventFollow
	if err := json.Unmarshal(m, &payload); err != nil {
		return fmt.Errorf("unmarshalling event: %w", err)
	}

	evt := event.Follow{
		Channel:    "#" + payload.BroadcasterUserLogin,
		FollowedAt: payload.FollowedAt,
		User:       payload.UserLogin,
		UserID:     payload.UserID,
	}

	log.WithFields(event.ToLogFields(evt)).Info("User followed")
	go handleTypedMessage(ircHdl.Client(), nil, evt)

	return nil
}

func (*twitchWatcher) handleEventSubChannelOutboundRaid(m json.RawMessage) error {
	var payload twitch.EventSubEventRaid
	if err := json.Unmarshal(m, &payload); err != nil {
		return fmt.Errorf("unmarshalling event: %w", err)
	}

	evt := event.OutboundRaid{
		Channel: "#" + payload.FromBroadcasterUserLogin,
		To:      payload.ToBroadcasterUserLogin,
		ToID:    payload.ToBroadcasterUserID,
		Viewers: payload.Viewers,
	}

	log.WithFields(event.ToLogFields(evt)).Info("Outbound raid detected")
	go handleTypedMessage(ircHdl.Client(), nil, evt)

	return nil
}

func (*twitchWatcher) handleEventSubChannelPointCustomRewardRedemptionAdd(m json.RawMessage) error {
	var payload twitch.EventSubEventChannelPointCustomRewardRedemptionAdd
	if err := json.Unmarshal(m, &payload); err != nil {
		return fmt.Errorf("unmarshalling event: %w", err)
	}

	evt := event.ChannelPointRedeem{
		Channel:     "#" + payload.BroadcasterUserLogin,
		RewardCost:  payload.Reward.Cost,
		RewardID:    payload.Reward.ID,
		RewardTitle: payload.Reward.Title,
		Status:      payload.Status,
		User:        payload.UserLogin,
		UserID:      payload.UserID,
		UserInput:   payload.UserInput,
	}

	log.WithFields(event.ToLogFields(evt)).Info("ChannelPoint reward was redeemed")
	go handleTypedMessage(ircHdl.Client(), nil, evt)

	return nil
}

func (*twitchWatcher) handleEventSubChannelPollChange(eventType string) func(json.RawMessage) error {
	return func(m json.RawMessage) error {
		var payload twitch.EventSubEventPoll
		if err := json.Unmarshal(m, &payload); err != nil {
			return fmt.Errorf("unmarshalling event: %w", err)
		}

		poll := event.Poll{
			Channel:               "#" + payload.BroadcasterUserLogin,
			HasChannelPointVoting: payload.ChannelPointsVoting.IsEnabled,
			Poll:                  payload,
			Title:                 payload.Title,
		}

		var evt event.Event

		switch eventType {
		case *event.PollBegin{}.Event():
			evt = event.PollBegin{
				Poll: poll,
			}

			log.WithFields(event.ToLogFields(evt)).Info("Poll started")

		case *event.PollEnd{}.Event():
			evt = event.PollEnd{
				Poll:   poll,
				Status: payload.Status,
			}

			log.WithFields(event.ToLogFields(evt)).Info("Poll ended")

		case *event.PollProgress{}.Event():
			evt = event.PollProgress{
				Poll: poll,
			}

			// Lets not spam the info-level-log with every single vote but
			// provide them for bots with debug-level-logging
			log.WithFields(event.ToLogFields(evt)).Debug("Poll changed")
		}

		go handleTypedMessage(ircHdl.Client(), nil, evt)
		return nil
	}
}

func (t *twitchWatcher) handleEventSubChannelUpdate(m json.RawMessage) error {
	var payload twitch.EventSubEventChannelUpdate
	if err := json.Unmarshal(m, &payload); err != nil {
		return fmt.Errorf("unmarshalling event: %w", err)
	}

	t.triggerUpdate(payload.BroadcasterUserLogin, &payload.Title, &payload.CategoryName, nil)

	return nil
}

func (*twitchWatcher) handleEventSubHypetrainEvent(eventType string) func(json.RawMessage) error {
	return func(m json.RawMessage) error {
		var payload twitch.EventSubEventHypetrain
		if err := json.Unmarshal(m, &payload); err != nil {
			return fmt.Errorf("unmarshalling event: %w", err)
		}

		hypetrain := event.Hypetrain{
			Channel: "#" + payload.BroadcasterUserLogin,
			Event:   payload,
			Level:   payload.Level,
		}

		var levelProgress float64
		if payload.Goal > 0 {
			levelProgress = float64(payload.Progress) / float64(payload.Goal)
		}

		var evt event.Event
		switch eventType {
		case *event.HypetrainBegin{}.Event():
			evt = event.HypetrainBegin{
				Hypetrain:     hypetrain,
				LevelProgress: levelProgress,
			}

		case *event.HypetrainEnd{}.Event():
			evt = event.HypetrainEnd{
				Hypetrain: hypetrain,
			}

		case *event.HypetrainProgress{}.Event():
			evt = event.HypetrainProgress{
				Hypetrain:     hypetrain,
				LevelProgress: levelProgress,
			}
		}

		log.WithFields(event.ToLogFields(evt)).Info("Hypetrain event")
		go handleTypedMessage(ircHdl.Client(), nil, evt)

		return nil
	}
}

func (*twitchWatcher) handleEventSubModeratorAdd(m json.RawMessage) error {
	var payload twitch.EventSubEventUserRoleChange
	if err := json.Unmarshal(m, &payload); err != nil {
		return fmt.Errorf("unmarshalling event: %w", err)
	}

	evt := event.ModeratorAdd{
		Channel: "#" + payload.BroadcasterUserLogin,
		User:    payload.UserLogin,
		UserID:  payload.UserID,
	}

	log.WithFields(event.ToLogFields(evt)).Info("Moderator added")
	go handleTypedMessage(ircHdl.Client(), nil, evt)

	return nil
}

func (*twitchWatcher) handleEventSubModeratorRemove(m json.RawMessage) error {
	var payload twitch.EventSubEventUserRoleChange
	if err := json.Unmarshal(m, &payload); err != nil {
		return fmt.Errorf("unmarshalling event: %w", err)
	}

	evt := event.ModeratorRemove{
		Channel: "#" + payload.BroadcasterUserLogin,
		User:    payload.UserLogin,
		UserID:  payload.UserID,
	}

	log.WithFields(event.ToLogFields(evt)).Info("Moderator removed")
	go handleTypedMessage(ircHdl.Client(), nil, evt)

	return nil
}

func (*twitchWatcher) handleEventSubShoutoutCreated(m json.RawMessage) error {
	var payload twitch.EventSubEventShoutoutCreated
	if err := json.Unmarshal(m, &payload); err != nil {
		return fmt.Errorf("unmarshalling event: %w", err)
	}

	evt := event.ShoutoutCreated{
		Channel: "#" + payload.BroadcasterUserLogin,
		To:      payload.ToBroadcasterUserLogin,
		ToID:    payload.ToBroadcasterUserID,
		Viewers: payload.ViewerCount,
	}

	log.WithFields(event.ToLogFields(evt)).Info("Shoutout created")
	go handleTypedMessage(ircHdl.Client(), nil, evt)

	return nil
}

func (*twitchWatcher) handleEventSubShoutoutReceived(m json.RawMessage) error {
	var payload twitch.EventSubEventShoutoutReceived
	if err := json.Unmarshal(m, &payload); err != nil {
		return fmt.Errorf("unmarshalling event: %w", err)
	}

	evt := event.ShoutoutReceived{
		Channel: "#" + payload.BroadcasterUserLogin,
		From:    payload.FromBroadcasterUserLogin,
		FromID:  payload.FromBroadcasterUserID,
		Viewers: payload.ViewerCount,
	}

	log.WithFields(event.ToLogFields(evt)).Info("Shoutout received")
	go handleTypedMessage(ircHdl.Client(), nil, evt)

	return nil
}

func (t *twitchWatcher) handleEventSubStreamOnOff(isOnline bool) func(json.RawMessage) error {
	return func(m json.RawMessage) error {
		var payload twitch.EventSubEventFollow
		if err := json.Unmarshal(m, &payload); err != nil {
			return fmt.Errorf("unmarshalling event: %w", err)
		}

		t.triggerUpdate(payload.BroadcasterUserLogin, nil, nil, &isOnline)
		return nil
	}
}

func (*twitchWatcher) handleEventSubSusUserMessage(m json.RawMessage) (err error) {
	var payload twitch.EventSubEventSuspiciousUserMessage
	if err := json.Unmarshal(m, &payload); err != nil {
		return fmt.Errorf("unmarshalling event: %w", err)
	}

	evt := event.SuspiciousUserMessage{
		BanEvasion:        payload.BanEvasionEvaluation,
		Channel:           "#" + payload.BroadcasterUserLogin,
		Message:           payload.Message.Text,
		SharedBanChannels: payload.SharedBanChannelIDs,
		Status:            payload.LowTrustStatus,
		UserID:            payload.UserID,
		Username:          payload.UserLogin,
		UserType:          payload.Types,
	}

	log.WithFields(event.ToLogFields(evt)).Info("restricted user message")
	go handleTypedMessage(ircHdl.Client(), nil, evt)

	return nil
}

func (*twitchWatcher) handleEventSubSusUserUpdate(m json.RawMessage) (err error) {
	var payload twitch.EventSubEventSuspiciousUserUpdated
	if err := json.Unmarshal(m, &payload); err != nil {
		return fmt.Errorf("unmarshalling event: %w", err)
	}

	evt := event.SuspiciousUserUpdate{
		Channel:   "#" + payload.BroadcasterUserLogin,
		Moderator: payload.ModeratorUserLogin,
		Status:    payload.LowTrustStatus,
		UserID:    payload.UserID,
		Username:  payload.UserLogin,
	}

	log.WithFields(event.ToLogFields(evt)).Info("user restriction updated")
	go handleTypedMessage(ircHdl.Client(), nil, evt)

	return nil
}

func (*twitchWatcher) handleEventSubVIPAdd(m json.RawMessage) error {
	var payload twitch.EventSubEventUserRoleChange
	if err := json.Unmarshal(m, &payload); err != nil {
		return fmt.Errorf("unmarshalling event: %w", err)
	}

	evt := event.VIPAdd{
		Channel: "#" + payload.BroadcasterUserLogin,
		User:    payload.UserLogin,
		UserID:  payload.UserID,
	}

	log.WithFields(event.ToLogFields(evt)).Info("VIP added")
	go handleTypedMessage(ircHdl.Client(), nil, evt)

	return nil
}

func (*twitchWatcher) handleEventSubVIPRemove(m json.RawMessage) error {
	var payload twitch.EventSubEventUserRoleChange
	if err := json.Unmarshal(m, &payload); err != nil {
		return fmt.Errorf("unmarshalling event: %w", err)
	}

	evt := event.VIPRemove{
		Channel: "#" + payload.BroadcasterUserLogin,
		User:    payload.UserLogin,
		UserID:  payload.UserID,
	}

	log.WithFields(event.ToLogFields(evt)).Info("VIP removed")
	go handleTypedMessage(ircHdl.Client(), nil, evt)

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

		return nil, fmt.Errorf("getting twitch client for channel: %w", err)
	}

	userID, err := twitchClient.GetIDForUsername(context.Background(), channel)
	if err != nil {
		return nil, fmt.Errorf("resolving channel to user-id: %w", err)
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
				return nil, fmt.Errorf("checking granted scopes: %w", err)
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
		return nil, fmt.Errorf("getting eventsub client for channel: %w", err)
	}

	return esClient, nil
}

func (t *twitchWatcher) triggerUpdate(channel string, title, category *string, online *bool) {
	if category != nil && t.ChannelStatus[channel].Category != *category {
		t.ChannelStatus[channel].Category = *category

		evt := event.CategoryUpdate{
			Category: *category,
			Channel:  "#" + channel,
		}

		log.WithFields(event.ToLogFields(evt)).Info("Category updated")
		go handleTypedMessage(ircHdl.Client(), nil, evt)
	}

	if title != nil && t.ChannelStatus[channel].Title != *title {
		t.ChannelStatus[channel].Title = *title

		evt := event.TitleUpdate{
			Channel: "#" + channel,
			Title:   *title,
		}

		log.WithFields(event.ToLogFields(evt)).Info("Title updated")
		go handleTypedMessage(ircHdl.Client(), nil, evt)
	}

	if online != nil && t.ChannelStatus[channel].IsLive != *online {
		t.ChannelStatus[channel].IsLive = *online

		log.WithFields(log.Fields{
			"channel": channel,
			"isLive":  *online,
		}).Info("Live-status updated")

		var evt event.Event = event.StreamOnline{Channel: "#" + channel}
		if !*online {
			evt = event.StreamOffline{Channel: "#" + channel}
		}

		go handleTypedMessage(ircHdl.Client(), nil, evt)
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

	status.IsLive, err = twitchClient.HasLiveStream(context.Background(), channel)
	if err != nil {
		return fmt.Errorf("getting live status: %w", err)
	}

	status.Category, status.Title, err = twitchClient.GetRecentStreamInfo(context.Background(), channel)
	if err != nil {
		return fmt.Errorf("getting stream info: %w", err)
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
		return fmt.Errorf("registering eventsub callbacks: %w", err)
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
