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

	"github.com/gofrs/uuid/v3"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/go_helpers/v2/str"
)

const (
	eventSubHeaderMessageID        = "Twitch-Eventsub-Message-Id"
	eventSubHeaderMessageType      = "Twitch-Eventsub-Message-Type"
	eventSubHeaderMessageSignature = "Twitch-Eventsub-Message-Signature"
	eventSubHeaderMessageTimestamp = "Twitch-Eventsub-Message-Timestamp"

	eventSubMessageTypeVerification = "webhook_callback_verification"
	eventSubMessageTypeRevokation   = "revocation"

	eventSubStatusEnabled             = "enabled"
	eventSubStatusVerificationPending = "webhook_callback_verification_pending"
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
)

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
