package twitch

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid/v3"
	"github.com/gorilla/mux"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/go_helpers/v2/str"
)

const (
	eventSubHeaderMessageID        = "Twitch-Eventsub-Message-Id"
	eventSubHeaderMessageType      = "Twitch-Eventsub-Message-Type"
	eventSubHeaderMessageSignature = "Twitch-Eventsub-Message-Signature"
	eventSubHeaderMessageTimestamp = "Twitch-Eventsub-Message-Timestamp"
	// eventSubHeaderMessageRetry        = "Twitch-Eventsub-Message-Retry"
	// eventSubHeaderSubscriptionType    = "Twitch-Eventsub-Subscription-Type"
	// eventSubHeaderSubscriptionVersion = "Twitch-Eventsub-Subscription-Version"

	eventSubMessageTypeVerification = "webhook_callback_verification"
	eventSubMessageTypeRevokation   = "revocation"
	// eventSubMessageTypeNotification = "notification"

	eventSubStatusEnabled             = "enabled"
	eventSubStatusVerificationPending = "webhook_callback_verification_pending"
	// eventSubStatusAuthorizationRevoked = "authorization_revoked"
	// eventSubStatusFailuresExceeded     = "notification_failures_exceeded"
	// eventSubStatusUserRemoved          = "user_removed"
	// eventSubStatusVerificationFailed   = "webhook_callback_verification_failed"

	EventSubEventTypeChannelFollow                         = "channel.follow"
	EventSubEventTypeChannelPointCustomRewardRedemptionAdd = "channel.channel_points_custom_reward_redemption.add"
	EventSubEventTypeChannelRaid                           = "channel.raid"
	EventSubEventTypeChannelShoutoutCreate                 = "channel.shoutout.create"
	EventSubEventTypeChannelShoutoutReceive                = "channel.shoutout.receive"
	EventSubEventTypeChannelUpdate                         = "channel.update"
	EventSubEventTypeStreamOffline                         = "stream.offline"
	EventSubEventTypeStreamOnline                          = "stream.online"
	EventSubEventTypeUserAuthorizationRevoke               = "user.authorization.revoke"

	EventSubTopicVersion1    = "1"
	EventSubTopicVersion2    = "2"
	EventSubTopicVersionBeta = "beta"
)

type (
	// Deprecated: This client should no longer be used and will not be
	// maintained afterwards. Replace with EventSubSocketClient.
	EventSubClient struct {
		apiURL       string
		secret       string
		secretHandle string

		twitchClient *Client

		subscriptions     map[string]*registeredSubscription
		subscriptionsLock sync.RWMutex
	}

	EventSubCondition struct {
		BroadcasterUserID     string `json:"broadcaster_user_id,omitempty"`
		CampaignID            string `json:"campaign_id,omitempty"`
		CategoryID            string `json:"category_id,omitempty"`
		ClientID              string `json:"client_id,omitempty"`
		ExtensionClientID     string `json:"extension_client_id,omitempty"`
		FromBroadcasterUserID string `json:"from_broadcaster_user_id,omitempty"`
		OrganizationID        string `json:"organization_id,omitempty"`
		RewardID              string `json:"reward_id,omitempty"`
		ToBroadcasterUserID   string `json:"to_broadcaster_user_id,omitempty"`
		UserID                string `json:"user_id,omitempty"`
		ModeratorUserID       string `json:"moderator_user_id,omitempty"`
	}

	EventSubEventChannelPointCustomRewardRedemptionAdd struct {
		ID                   string `json:"id"`
		BroadcasterUserID    string `json:"broadcaster_user_id"`
		BroadcasterUserLogin string `json:"broadcaster_user_login"`
		BroadcasterUserName  string `json:"broadcaster_user_name"`
		UserID               string `json:"user_id"`
		UserLogin            string `json:"user_login"`
		UserName             string `json:"user_name"`
		UserInput            string `json:"user_input"`
		Status               string `json:"status"`
		Reward               struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Cost   int64  `json:"cost"`
			Prompt string `json:"prompt"`
		} `json:"reward"`
		RedeemedAt time.Time `json:"redeemed_at"`
	}

	EventSubEventChannelUpdate struct {
		BroadcasterUserID    string `json:"broadcaster_user_id"`
		BroadcasterUserLogin string `json:"broadcaster_user_login"`
		BroadcasterUserName  string `json:"broadcaster_user_name"`
		Title                string `json:"title"`
		Language             string `json:"language"`
		CategoryID           string `json:"category_id"`
		CategoryName         string `json:"category_name"`
		IsMature             bool   `json:"is_mature"`
	}

	EventSubEventFollow struct {
		UserID               string    `json:"user_id"`
		UserLogin            string    `json:"user_login"`
		UserName             string    `json:"user_name"`
		BroadcasterUserID    string    `json:"broadcaster_user_id"`
		BroadcasterUserLogin string    `json:"broadcaster_user_login"`
		BroadcasterUserName  string    `json:"broadcaster_user_name"`
		FollowedAt           time.Time `json:"followed_at"`
	}

	EventSubEventRaid struct {
		FromBroadcasterUserID    string `json:"from_broadcaster_user_id"`
		FromBroadcasterUserLogin string `json:"from_broadcaster_user_login"`
		FromBroadcasterUserName  string `json:"from_broadcaster_user_name"`
		ToBroadcasterUserID      string `json:"to_broadcaster_user_id"`
		ToBroadcasterUserLogin   string `json:"to_broadcaster_user_login"`
		ToBroadcasterUserName    string `json:"to_broadcaster_user_name"`
		Viewers                  int64  `json:"viewers"`
	}

	EventSubEventShoutoutCreated struct {
		BroadcasterUserID      string    `json:"broadcaster_user_id"`
		BroadcasterUserLogin   string    `json:"broadcaster_user_login"`
		BroadcasterUserName    string    `json:"broadcaster_user_name"`
		ModeratorUserID        string    `json:"moderator_user_id"`
		ModeratorUserLogin     string    `json:"moderator_user_login"`
		ModeratorUserName      string    `json:"moderator_user_name"`
		ToBroadcasterUserID    string    `json:"to_broadcaster_user_id"`
		ToBroadcasterUserLogin string    `json:"to_broadcaster_user_login"`
		ToBroadcasterUserName  string    `json:"to_broadcaster_user_name"`
		ViewerCount            int64     `json:"viewer_count"`
		StartedAt              time.Time `json:"started_at"`
		CooldownEndsAt         time.Time `json:"cooldown_ends_at"`
		TargetCooldownEndsAt   time.Time `json:"target_cooldown_ends_at"`
	}

	EventSubEventShoutoutReceived struct {
		BroadcasterUserID        string    `json:"broadcaster_user_id"`
		BroadcasterUserLogin     string    `json:"broadcaster_user_login"`
		BroadcasterUserName      string    `json:"broadcaster_user_name"`
		FromBroadcasterUserID    string    `json:"from_broadcaster_user_id"`
		FromBroadcasterUserLogin string    `json:"from_broadcaster_user_login"`
		FromBroadcasterUserName  string    `json:"from_broadcaster_user_name"`
		ViewerCount              int64     `json:"viewer_count"`
		StartedAt                time.Time `json:"started_at"`
	}

	EventSubEventStreamOffline struct {
		BroadcasterUserID    string `json:"broadcaster_user_id"`
		BroadcasterUserLogin string `json:"broadcaster_user_login"`
		BroadcasterUserName  string `json:"broadcaster_user_name"`
	}

	EventSubEventStreamOnline struct {
		ID                   string    `json:"id"`
		BroadcasterUserID    string    `json:"broadcaster_user_id"`
		BroadcasterUserLogin string    `json:"broadcaster_user_login"`
		BroadcasterUserName  string    `json:"broadcaster_user_name"`
		Type                 string    `json:"type"`
		StartedAt            time.Time `json:"started_at"`
	}

	EventSubEventUserAuthorizationRevoke struct {
		ClientID  string `json:"client_id"`
		UserID    string `json:"user_id"`
		UserLogin string `json:"user_login"`
		UserName  string `json:"user_name"`
	}

	eventSubPostMessage struct {
		Challenge    string               `json:"challenge"`
		Subscription eventSubSubscription `json:"subscription"`
		Event        json.RawMessage      `json:"event"`
	}

	eventSubSubscription struct {
		ID        string            `json:"id,omitempty"`     // READONLY
		Status    string            `json:"status,omitempty"` // READONLY
		Type      string            `json:"type"`
		Version   string            `json:"version"`
		Cost      int64             `json:"cost,omitempty"` // READONLY
		Condition EventSubCondition `json:"condition"`
		Transport eventSubTransport `json:"transport"`
		CreatedAt time.Time         `json:"created_at,omitempty"` // READONLY
	}

	eventSubTransport struct {
		Method    string `json:"method"`
		Callback  string `json:"callback"`
		Secret    string `json:"secret"`
		SessionID string `json:"session_id"`
	}

	registeredSubscription struct {
		Type         string
		Callbacks    map[string]func(json.RawMessage) error
		Subscription eventSubSubscription
	}
)

func (e EventSubCondition) Hash() (string, error) {
	h, err := hashstructure.Hash(e, hashstructure.FormatV2, &hashstructure.HashOptions{TagName: "json"})
	if err != nil {
		return "", errors.Wrap(err, "hashing struct")
	}

	return fmt.Sprintf("%x", h), nil
}

// Deprecated: See deprecation notice of EventSubClient
func NewEventSubClient(twitchClient *Client, apiURL, secret, secretHandle string) (*EventSubClient, error) {
	c := &EventSubClient{
		apiURL:       apiURL,
		secret:       secret,
		secretHandle: secretHandle,

		twitchClient: twitchClient,

		subscriptions: map[string]*registeredSubscription{},
	}

	return c, c.PreFetchSubscriptions(context.Background())
}

func (e *EventSubClient) HandleEventsubPush(w http.ResponseWriter, r *http.Request) {
	var (
		body      = new(bytes.Buffer)
		keyHandle = mux.Vars(r)["keyhandle"]
		message   eventSubPostMessage
		signature = r.Header.Get(eventSubHeaderMessageSignature)
	)

	if keyHandle != e.secretHandle {
		http.Error(w, "deprecated callback used", http.StatusNotFound)
		return
	}

	// Copy body for duplicate processing
	if _, err := io.Copy(body, r.Body); err != nil {
		log.WithError(err).Error("Unable to read hook body")
		return
	}

	// Verify signature
	mac := hmac.New(sha256.New, []byte(e.secret))
	fmt.Fprintf(mac, "%s%s%s", r.Header.Get(eventSubHeaderMessageID), r.Header.Get(eventSubHeaderMessageTimestamp), body.Bytes())
	if cSig := fmt.Sprintf("sha256=%x", mac.Sum(nil)); cSig != signature {
		log.Errorf("Got message signature %s, expected %s", signature, cSig)
		http.Error(w, "Signature verification failed", http.StatusUnauthorized)
		return
	}

	// Read message
	if err := json.NewDecoder(body).Decode(&message); err != nil {
		log.WithError(err).Errorf("Unable to decode eventsub message")
		http.Error(w, errors.Wrap(err, "parsing message").Error(), http.StatusBadRequest)
		return
	}

	logger := log.WithField("type", message.Subscription.Type)

	// If we got a verification request, respond with the challenge
	switch r.Header.Get(eventSubHeaderMessageType) {
	case eventSubMessageTypeRevokation:
		w.WriteHeader(http.StatusNoContent)
		return

	case eventSubMessageTypeVerification:
		logger.Debug("Confirming eventsub subscription")
		w.Write([]byte(message.Challenge))
		return
	}

	logger.Debug("Received notification")

	condHash, err := message.Subscription.Condition.Hash()
	if err != nil {
		logger.WithError(err).Errorf("Unable to hash condition of push")
		http.Error(w, errors.Wrap(err, "hashing condition").Error(), http.StatusBadRequest)
		return
	}

	e.subscriptionsLock.RLock()
	defer e.subscriptionsLock.RUnlock()

	cacheKey := strings.Join([]string{message.Subscription.Type, message.Subscription.Version, condHash}, "::")

	reg, ok := e.subscriptions[cacheKey]
	if !ok {
		http.Error(w, "no subscription available", http.StatusBadRequest)
		return
	}

	for _, cb := range reg.Callbacks {
		if err = cb(message.Event); err != nil {
			logger.WithError(err).Error("Handler callback caused error")
		}
	}
}

func (e *EventSubClient) PreFetchSubscriptions(ctx context.Context) error {
	e.subscriptionsLock.Lock()
	defer e.subscriptionsLock.Unlock()

	subList, err := e.twitchClient.getEventSubSubscriptions(ctx)
	if err != nil {
		return errors.Wrap(err, "listing existing subscriptions")
	}

	for i := range subList {
		sub := subList[i]

		switch {
		case !str.StringInSlice(sub.Status, []string{eventSubStatusEnabled, eventSubStatusVerificationPending}):
			// Is not an active hook, we don't need to care: It will be
			// confirmed later or will expire but should not be counted
			continue

		case strings.HasPrefix(sub.Transport.Callback, e.apiURL) && sub.Transport.Callback != e.fullAPIurl():
			// Uses the same API URL but with another secret handle: Must
			// have been registered by another instance with another secret
			// so we should be able to deregister it without causing any
			// trouble
			logger := log.WithFields(log.Fields{
				"id":      sub.ID,
				"topic":   sub.Type,
				"version": sub.Version,
			})
			logger.Debug("Removing deprecated EventSub subscription")
			if err = e.twitchClient.deleteEventSubSubscription(ctx, sub.ID); err != nil {
				logger.WithError(err).Error("Unable to deregister deprecated EventSub subscription")
			}
			continue

		case sub.Transport.Callback != e.fullAPIurl():
			// Different callback URL: We don't care, it's probably another
			// bot instance with the same client ID
			continue
		}

		condHash, err := sub.Condition.Hash()
		if err != nil {
			return errors.Wrap(err, "hashing condition")
		}

		log.WithFields(log.Fields{
			"condition": sub.Condition,
			"type":      sub.Type,
			"version":   sub.Version,
		}).Debug("found existing eventsub subscription")

		cacheKey := strings.Join([]string{sub.Type, sub.Version, condHash}, "::")
		e.subscriptions[cacheKey] = &registeredSubscription{
			Type:         sub.Type,
			Callbacks:    map[string]func(json.RawMessage) error{},
			Subscription: sub,
		}
	}

	return nil
}

func (e *EventSubClient) RegisterEventSubHooks(event, version string, condition EventSubCondition, callback func(json.RawMessage) error) (func(), error) {
	if version == "" {
		version = EventSubTopicVersion1
	}

	condHash, err := condition.Hash()
	if err != nil {
		return nil, errors.Wrap(err, "hashing condition")
	}

	var (
		cacheKey = strings.Join([]string{event, version, condHash}, "::")
		logger   = log.WithFields(log.Fields{
			"condition": condition,
			"type":      event,
			"version":   version,
		})
	)

	e.subscriptionsLock.RLock()
	_, ok := e.subscriptions[cacheKey]
	e.subscriptionsLock.RUnlock()

	if ok {
		// Subscription already exists
		e.subscriptionsLock.Lock()
		defer e.subscriptionsLock.Unlock()

		logger.Debug("Adding callback to known subscription")

		cbKey := uuid.Must(uuid.NewV4()).String()

		e.subscriptions[cacheKey].Callbacks[cbKey] = callback
		return func() { e.unregisterCallback(cacheKey, cbKey) }, nil
	}

	logger.Debug("registering new eventsub subscription")

	// Register subscriptions
	ctx, cancel := context.WithTimeout(context.Background(), twitchRequestTimeout)
	defer cancel()

	newSub, err := e.twitchClient.createEventSubSubscriptionWebhook(ctx, eventSubSubscription{
		Type:      event,
		Version:   version,
		Condition: condition,
		Transport: eventSubTransport{
			Method:   "webhook",
			Callback: e.fullAPIurl(),
			Secret:   e.secret,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating subscription")
	}

	e.subscriptionsLock.Lock()
	defer e.subscriptionsLock.Unlock()

	logger.Debug("Registered new hook")

	cbKey := uuid.Must(uuid.NewV4()).String()
	e.subscriptions[cacheKey] = &registeredSubscription{
		Type: event,
		Callbacks: map[string]func(json.RawMessage) error{
			cbKey: callback,
		},
		Subscription: *newSub,
	}

	logger.Debug("Registered eventsub subscription")

	return func() { e.unregisterCallback(cacheKey, cbKey) }, nil
}

func (c *Client) createEventSubSubscriptionWebhook(ctx context.Context, sub eventSubSubscription) (*eventSubSubscription, error) {
	return c.createEventSubSubscription(ctx, authTypeAppAccessToken, sub)
}

func (c *Client) createEventSubSubscriptionWebsocket(ctx context.Context, sub eventSubSubscription) (*eventSubSubscription, error) {
	return c.createEventSubSubscription(ctx, authTypeBearerToken, sub)
}

func (c *Client) createEventSubSubscription(ctx context.Context, auth authType, sub eventSubSubscription) (*eventSubSubscription, error) {
	var (
		buf  = new(bytes.Buffer)
		resp struct {
			Total      int64                  `json:"total"`
			Data       []eventSubSubscription `json:"data"`
			Pagination struct {
				Cursor string `json:"cursor"`
			} `json:"pagination"`
		}
	)

	if err := json.NewEncoder(buf).Encode(sub); err != nil {
		return nil, errors.Wrap(err, "assemble subscribe payload")
	}

	if err := c.request(clientRequestOpts{
		AuthType: auth,
		Body:     buf,
		Context:  ctx,
		Method:   http.MethodPost,
		OKStatus: http.StatusAccepted,
		Out:      &resp,
		URL:      "https://api.twitch.tv/helix/eventsub/subscriptions",
	}); err != nil {
		return nil, errors.Wrap(err, "executing request")
	}

	return &resp.Data[0], nil
}

func (c *Client) deleteEventSubSubscription(ctx context.Context, id string) error {
	return errors.Wrap(c.request(clientRequestOpts{
		AuthType: authTypeAppAccessToken,
		Context:  ctx,
		Method:   http.MethodDelete,
		OKStatus: http.StatusNoContent,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/eventsub/subscriptions?id=%s", id),
	}), "executing request")
}

func (e *EventSubClient) fullAPIurl() string {
	return strings.Join([]string{e.apiURL, e.secretHandle}, "/")
}

func (c *Client) getEventSubSubscriptions(ctx context.Context) ([]eventSubSubscription, error) {
	var (
		out    []eventSubSubscription
		params = make(url.Values)
		resp   struct {
			Total      int64                  `json:"total"`
			Data       []eventSubSubscription `json:"data"`
			Pagination struct {
				Cursor string `json:"cursor"`
			} `json:"pagination"`
		}
	)

	for {
		if err := c.request(clientRequestOpts{
			AuthType: authTypeAppAccessToken,
			Context:  ctx,
			Method:   http.MethodGet,
			OKStatus: http.StatusOK,
			Out:      &resp,
			URL:      fmt.Sprintf("https://api.twitch.tv/helix/eventsub/subscriptions?%s", params.Encode()),
		}); err != nil {
			return nil, errors.Wrap(err, "executing request")
		}

		out = append(out, resp.Data...)

		if resp.Pagination.Cursor == "" {
			break
		}

		params.Set("after", resp.Pagination.Cursor)

		// Clear from struct as struct is reused
		resp.Data = nil
		resp.Pagination.Cursor = ""
	}

	return out, nil
}

func (e *EventSubClient) unregisterCallback(cacheKey, cbKey string) {
	e.subscriptionsLock.RLock()
	regSub, ok := e.subscriptions[cacheKey]
	e.subscriptionsLock.RUnlock()

	if !ok {
		// That subscription does not exist
		log.WithField("cache_key", cacheKey).Debug("Subscription does not exist, not unregistering")
		return
	}

	if _, ok = regSub.Callbacks[cbKey]; !ok {
		// That callback does not exist
		log.WithFields(log.Fields{
			"cache_key": cacheKey,
			"callback":  cbKey,
		}).Debug("Callback does not exist, not unregistering")
		return
	}

	logger := log.WithField("event", regSub.Type)

	delete(regSub.Callbacks, cbKey)

	if len(regSub.Callbacks) > 0 {
		// Still callbacks registered, not removing the subscription
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), twitchRequestTimeout)
	defer cancel()

	if err := e.twitchClient.deleteEventSubSubscription(ctx, regSub.Subscription.ID); err != nil {
		log.WithError(err).Error("Unable to execute delete subscription request")
		return
	}

	e.subscriptionsLock.Lock()
	defer e.subscriptionsLock.Unlock()

	logger.Debug("Unregistered hook")

	delete(e.subscriptions, cacheKey)
}
