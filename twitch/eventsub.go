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

	EventSubEventTypeChannelFollow = "channel.follow"
	EventSubEventTypeChannelUpdate = "channel.update"
	EventSubEventTypeStreamOffline = "stream.offline"
	EventSubEventTypeStreamOnline  = "stream.online"
)

type (
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
		Method   string `json:"method"`
		Callback string `json:"callback"`
		Secret   string `json:"secret"`
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

func NewEventSubClient(twitchClient *Client, apiURL, secret, secretHandle string) *EventSubClient {
	return &EventSubClient{
		apiURL:       apiURL,
		secret:       secret,
		secretHandle: secretHandle,

		twitchClient: twitchClient,

		subscriptions: map[string]*registeredSubscription{},
	}
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

	cacheKey := strings.Join([]string{message.Subscription.Type, condHash}, "::")

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

//nolint:funlen // Not splitting to keep logic unit
func (e *EventSubClient) RegisterEventSubHooks(event string, condition EventSubCondition, callback func(json.RawMessage) error) (func(), error) {
	condHash, err := condition.Hash()
	if err != nil {
		return nil, errors.Wrap(err, "hashing condition")
	}

	var (
		cacheKey = strings.Join([]string{event, condHash}, "::")
		logger   = log.WithField("event", event)
	)

	e.subscriptionsLock.RLock()
	_, ok := e.subscriptions[cacheKey]
	e.subscriptionsLock.RUnlock()

	if ok {
		// Subscription already exists
		e.subscriptionsLock.Lock()
		defer e.subscriptionsLock.Unlock()

		logger.Debug("Adding callback to existing callback")

		cbKey := uuid.Must(uuid.NewV4()).String()

		e.subscriptions[cacheKey].Callbacks[cbKey] = callback
		return func() { e.unregisterCallback(cacheKey, cbKey) }, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), twitchRequestTimeout)
	defer cancel()

	// List existing subscriptions
	subList, err := e.twitchClient.getEventSubSubscriptions(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "listing existing subscriptions")
	}

	// Register subscriptions
	var (
		existingSub *eventSubSubscription
	)
	for i, sub := range subList {
		existingConditionHash, err := sub.Condition.Hash()
		if err != nil {
			return nil, errors.Wrap(err, "hashing existing condition")
		}
		newConditionHash, err := condition.Hash()
		if err != nil {
			return nil, errors.Wrap(err, "hashing new condition")
		}

		if str.StringInSlice(sub.Status, []string{eventSubStatusEnabled, eventSubStatusVerificationPending}) &&
			sub.Transport.Callback == e.apiURL &&
			existingConditionHash == newConditionHash &&
			sub.Type == event {
			logger = logger.WithFields(log.Fields{
				"id":     sub.ID,
				"status": sub.Status,
			})
			existingSub = &subList[i]
		}
	}

	if existingSub != nil {
		logger.WithField("event", event).Debug("Not registering hook, already active")

		e.subscriptionsLock.Lock()
		defer e.subscriptionsLock.Unlock()

		logger.Debug("Found existing hook, registering and adding callback")

		cbKey := uuid.Must(uuid.NewV4()).String()
		e.subscriptions[cacheKey] = &registeredSubscription{
			Type: event,
			Callbacks: map[string]func(json.RawMessage) error{
				cbKey: callback,
			},
			Subscription: *existingSub,
		}

		return func() { e.unregisterCallback(cacheKey, cbKey) }, nil
	}

	ctx, cancel = context.WithTimeout(context.Background(), twitchRequestTimeout)
	defer cancel()

	newSub, err := e.twitchClient.createEventSubSubscription(ctx, eventSubSubscription{
		Type:      event,
		Version:   "1",
		Condition: condition,
		Transport: eventSubTransport{
			Method:   "webhook",
			Callback: e.apiURL,
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
