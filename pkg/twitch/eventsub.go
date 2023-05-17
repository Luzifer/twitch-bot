package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
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
