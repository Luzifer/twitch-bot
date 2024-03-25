package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/pkg/errors"
)

// Collection of known EventSub event-types
const (
	EventSubEventTypeChannelAdBreakBegin                   = "channel.ad_break.begin"
	EventSubEventTypeChannelFollow                         = "channel.follow"
	EventSubEventTypeChannelPointCustomRewardRedemptionAdd = "channel.channel_points_custom_reward_redemption.add"
	EventSubEventTypeChannelHypetrainBegin                 = "channel.hype_train.begin"
	EventSubEventTypeChannelHypetrainProgress              = "channel.hype_train.progress"
	EventSubEventTypeChannelHypetrainEnd                   = "channel.hype_train.end"
	EventSubEventTypeChannelRaid                           = "channel.raid"
	EventSubEventTypeChannelShoutoutCreate                 = "channel.shoutout.create"
	EventSubEventTypeChannelShoutoutReceive                = "channel.shoutout.receive"
	EventSubEventTypeChannelUpdate                         = "channel.update"
	EventSubEventTypeChannelPollBegin                      = "channel.poll.begin"
	EventSubEventTypeChannelPollEnd                        = "channel.poll.end"
	EventSubEventTypeChannelPollProgress                   = "channel.poll.progress"
	EventSubEventTypeStreamOffline                         = "stream.offline"
	EventSubEventTypeStreamOnline                          = "stream.online"
	EventSubEventTypeUserAuthorizationRevoke               = "user.authorization.revoke"
)

// Collection of topic versions known to the API
const (
	EventSubTopicVersion1    = "1"
	EventSubTopicVersion2    = "2"
	EventSubTopicVersionBeta = "beta"
)

type (
	// EventSubCondition defines the condition the subscription should
	// listen on - all fields are optional and those defined in the
	// EventSub documentation for the given topic should be set
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

	// EventSubEventAdBreakBegin contains the payload for an AdBreak event
	EventSubEventAdBreakBegin struct {
		Duration             int64     `json:"duration_seconds"`
		StartedAt            time.Time `json:"started_at"`
		IsAutomatic          bool      `json:"is_automatic"`
		BroadcasterUserID    string    `json:"broadcaster_user_id"`
		BroadcasterUserLogin string    `json:"broadcaster_user_login"`
		BroadcasterUserName  string    `json:"broadcaster_user_name"`
		RequesterUserID      string    `json:"requester_user_id"`
		RequesterUserLogin   string    `json:"requester_user_login"`
		RequesterUserName    string    `json:"requester_user_name"`
	}

	// EventSubEventChannelPointCustomRewardRedemptionAdd contains the
	// payload for an channel-point redeem event
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

	// EventSubEventChannelUpdate contains the payload for a channel
	// update event
	EventSubEventChannelUpdate struct {
		BroadcasterUserID           string   `json:"broadcaster_user_id"`
		BroadcasterUserLogin        string   `json:"broadcaster_user_login"`
		BroadcasterUserName         string   `json:"broadcaster_user_name"`
		Title                       string   `json:"title"`
		Language                    string   `json:"language"`
		CategoryID                  string   `json:"category_id"`
		CategoryName                string   `json:"category_name"`
		ContentClassificationLabels []string `json:"content_classification_labels"`
	}

	// EventSubEventFollow contains the payload for a follow event
	EventSubEventFollow struct {
		UserID               string    `json:"user_id"`
		UserLogin            string    `json:"user_login"`
		UserName             string    `json:"user_name"`
		BroadcasterUserID    string    `json:"broadcaster_user_id"`
		BroadcasterUserLogin string    `json:"broadcaster_user_login"`
		BroadcasterUserName  string    `json:"broadcaster_user_name"`
		FollowedAt           time.Time `json:"followed_at"`
	}

	// EventSubEventHypetrain contains the payload for all three (begin,
	// progress and end) hypetrain events. Certain fields are not
	// available at all event types
	EventSubEventHypetrain struct {
		ID                   string `json:"id"`
		BroadcasterUserID    string `json:"broadcaster_user_id"`
		BroadcasterUserLogin string `json:"broadcaster_user_login"`
		BroadcasterUserName  string `json:"broadcaster_user_name"`
		Level                int64  `json:"level"`
		Total                int64  `json:"total"`
		Progress             int64  `json:"progress"` // Only Beginn, Progress
		Goal                 int64  `json:"goal"`     // Only Beginn, Progress
		TopContributions     []struct {
			UserID    string `json:"user_id"`
			UserLogin string `json:"user_login"`
			UserName  string `json:"user_name"`
			Type      string `json:"type"`
			Total     int64  `json:"total"`
		} `json:"top_contributions"`
		LastContribution *struct { // Only Begin, Progress
			UserID    string `json:"user_id"`
			UserLogin string `json:"user_login"`
			UserName  string `json:"user_name"`
			Type      string `json:"type"`
			Total     int64  `json:"total"`
		} `json:"last_contribution,omitempty"`
		StartedAt      time.Time  `json:"started_at"`
		ExpiresAt      *time.Time `json:"expires_at,omitempty"`       // Only Begin, Progress
		EndedAt        *time.Time `json:"ended_at,omitempty"`         // Only End
		CooldownEndsAt *time.Time `json:"cooldown_ends_at,omitempty"` // Only End
	}

	// EventSubEventPoll contains the payload for a poll change event
	// (not all fields are present in all poll events, see docs!)
	EventSubEventPoll struct {
		ID                   string `json:"id"`
		BroadcasterUserID    string `json:"broadcaster_user_id"`
		BroadcasterUserLogin string `json:"broadcaster_user_login"`
		BroadcasterUserName  string `json:"broadcaster_user_name"`
		Title                string `json:"title"`
		Choices              []struct {
			ID                 string `json:"id"`
			Title              string `json:"title"`
			ChannelPointsVotes int    `json:"channel_points_votes"`
			Votes              int    `json:"votes"`
		} `json:"choices"`
		ChannelPointsVoting struct {
			IsEnabled     bool `json:"is_enabled"`
			AmountPerVote int  `json:"amount_per_vote"`
		} `json:"channel_points_voting"`

		StartedAt time.Time `json:"started_at"`         // begin, progress, end
		EndsAt    time.Time `json:"ends_at,omitempty"`  // begin, progress
		Status    string    `json:"status,omitempty"`   // end -- enum(completed, archived, terminated)
		EndedAt   time.Time `json:"ended_at,omitempty"` // end
	}

	// EventSubEventRaid contains the payload for a raid event
	EventSubEventRaid struct {
		FromBroadcasterUserID    string `json:"from_broadcaster_user_id"`
		FromBroadcasterUserLogin string `json:"from_broadcaster_user_login"`
		FromBroadcasterUserName  string `json:"from_broadcaster_user_name"`
		ToBroadcasterUserID      string `json:"to_broadcaster_user_id"`
		ToBroadcasterUserLogin   string `json:"to_broadcaster_user_login"`
		ToBroadcasterUserName    string `json:"to_broadcaster_user_name"`
		Viewers                  int64  `json:"viewers"`
	}

	// EventSubEventShoutoutCreated contains the payload for a shoutout
	// created event
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

	// EventSubEventShoutoutReceived contains the payload for a shoutout
	// received event
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

	// EventSubEventStreamOffline contains the payload for a stream
	// offline event
	EventSubEventStreamOffline struct {
		BroadcasterUserID    string `json:"broadcaster_user_id"`
		BroadcasterUserLogin string `json:"broadcaster_user_login"`
		BroadcasterUserName  string `json:"broadcaster_user_name"`
	}

	// EventSubEventStreamOnline contains the payload for a stream
	// online event
	EventSubEventStreamOnline struct {
		ID                   string    `json:"id"`
		BroadcasterUserID    string    `json:"broadcaster_user_id"`
		BroadcasterUserLogin string    `json:"broadcaster_user_login"`
		BroadcasterUserName  string    `json:"broadcaster_user_name"`
		Type                 string    `json:"type"`
		StartedAt            time.Time `json:"started_at"`
	}

	// EventSubEventUserAuthorizationRevoke contains the payload for an
	// authorization revoke event
	EventSubEventUserAuthorizationRevoke struct {
		ClientID  string `json:"client_id"`
		UserID    string `json:"user_id"`
		UserLogin string `json:"user_login"`
		UserName  string `json:"user_name"`
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
)

// Hash generates a hashstructure hash for the condition for comparison
func (e EventSubCondition) Hash() (string, error) {
	h, err := hashstructure.Hash(e, hashstructure.FormatV2, &hashstructure.HashOptions{TagName: "json"})
	if err != nil {
		return "", errors.Wrap(err, "hashing struct")
	}

	return fmt.Sprintf("%x", h), nil
}

func (c *Client) createEventSubSubscriptionWebsocket(ctx context.Context, sub eventSubSubscription) (*eventSubSubscription, error) {
	return c.createEventSubSubscription(ctx, AuthTypeBearerToken, sub)
}

func (c *Client) createEventSubSubscription(ctx context.Context, auth AuthType, sub eventSubSubscription) (*eventSubSubscription, error) {
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

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: auth,
		Body:     buf,
		Method:   http.MethodPost,
		OKStatus: http.StatusAccepted,
		Out:      &resp,
		URL:      "https://api.twitch.tv/helix/eventsub/subscriptions",
		ValidateFunc: func(opts ClientRequestOpts, resp *http.Response) error {
			if resp.StatusCode == http.StatusConflict {
				// This is fine: We needed that subscription, it exists
				return nil
			}

			// Fallback to default handling
			return ValidateStatus(opts, resp)
		},
	}); err != nil {
		return nil, errors.Wrap(err, "executing request")
	}

	return &resp.Data[0], nil
}
