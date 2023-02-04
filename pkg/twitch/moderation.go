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

	"github.com/pkg/errors"
)

const maxTimeoutDuration = 1209600 * time.Second

// BanUser bans or timeouts a user in the given channel. Setting the
// duration to 0 will result in a ban, setting if greater than 0 will
// result in a timeout. The timeout is automatically converted to
// full seconds. The timeout duration must be less than 1209600s.
func (c *Client) BanUser(channel, username string, duration time.Duration, reason string) error {
	var payload struct {
		Data struct {
			Duration int64  `json:"duration,omitempty"`
			Reason   string `json:"reason"`
			UserID   string `json:"user_id"`
		} `json:"data"`
	}

	if duration > maxTimeoutDuration {
		return errors.New("timeout duration exceeds maximum")
	}

	payload.Data.Duration = int64(duration / time.Second)
	payload.Data.Reason = reason

	botID, _, err := c.GetAuthorizedUser()
	if err != nil {
		return errors.Wrap(err, "getting bot user-id")
	}

	channelID, err := c.GetIDForUsername(strings.TrimLeft(channel, "#@"))
	if err != nil {
		return errors.Wrap(err, "getting channel user-id")
	}

	if payload.Data.UserID, err = c.GetIDForUsername(username); err != nil {
		return errors.Wrap(err, "getting target user-id")
	}

	body := new(bytes.Buffer)
	if err = json.NewEncoder(body).Encode(payload); err != nil {
		return errors.Wrap(err, "encoding payload")
	}

	return errors.Wrap(
		c.request(clientRequestOpts{
			AuthType: authTypeBearerToken,
			Context:  context.Background(),
			Method:   http.MethodPost,
			OKStatus: http.StatusOK,
			Body:     body,
			URL: fmt.Sprintf(
				"https://api.twitch.tv/helix/moderation/bans?broadcaster_id=%s&moderator_id=%s",
				channelID, botID,
			),
		}),
		"executing ban request",
	)
}

// DeleteMessage deletes one or all messages from the specified chat.
// If no messageID is given all messages are deleted. If a message ID
// is given the message must be no older than 6 hours and it must not
// be posted by broadcaster or moderator.
func (c *Client) DeleteMessage(channel, messageID string) error {
	botID, _, err := c.GetAuthorizedUser()
	if err != nil {
		return errors.Wrap(err, "getting bot user-id")
	}

	channelID, err := c.GetIDForUsername(strings.TrimLeft(channel, "#@"))
	if err != nil {
		return errors.Wrap(err, "getting channel user-id")
	}

	params := make(url.Values)
	params.Set("broadcaster_id", channelID)
	params.Set("moderator_id", botID)
	if messageID != "" {
		params.Set("message_id", messageID)
	}

	return errors.Wrap(
		c.request(clientRequestOpts{
			AuthType: authTypeBearerToken,
			Context:  context.Background(),
			Method:   http.MethodDelete,
			OKStatus: http.StatusNoContent,
			URL: fmt.Sprintf(
				"https://api.twitch.tv/helix/moderation/chat?%s",
				params.Encode(),
			),
		}),
		"executing delete request",
	)
}

// UnbanUser removes a timeout or ban given to the user in the channel
func (c *Client) UnbanUser(channel, username string) error {
	botID, _, err := c.GetAuthorizedUser()
	if err != nil {
		return errors.Wrap(err, "getting bot user-id")
	}

	channelID, err := c.GetIDForUsername(strings.TrimLeft(channel, "#@"))
	if err != nil {
		return errors.Wrap(err, "getting channel user-id")
	}

	userID, err := c.GetIDForUsername(username)
	if err != nil {
		return errors.Wrap(err, "getting target user-id")
	}

	return errors.Wrap(
		c.request(clientRequestOpts{
			AuthType: authTypeBearerToken,
			Context:  context.Background(),
			Method:   http.MethodDelete,
			OKStatus: http.StatusNoContent,
			URL: fmt.Sprintf(
				"https://api.twitch.tv/helix/moderation/bans?broadcaster_id=%s&moderator_id=%s&user_id=%s",
				channelID, botID, userID,
			),
		}),
		"executing unban request",
	)
}

// UpdateShieldMode activates or deactivates the Shield Mode in the given channel
func (c *Client) UpdateShieldMode(ctx context.Context, channel string, enable bool) error {
	botID, _, err := c.GetAuthorizedUser()
	if err != nil {
		return errors.Wrap(err, "getting bot user-id")
	}

	channelID, err := c.GetIDForUsername(strings.TrimLeft(channel, "#@"))
	if err != nil {
		return errors.Wrap(err, "getting channel user-id")
	}

	body := new(bytes.Buffer)
	if err = json.NewEncoder(body).Encode(map[string]bool{
		"is_active": enable,
	}); err != nil {
		return errors.Wrap(err, "encoding payload")
	}

	return errors.Wrap(
		c.request(clientRequestOpts{
			AuthType: authTypeBearerToken,
			Context:  ctx,
			Method:   http.MethodPut,
			OKStatus: http.StatusOK,
			Body:     body,
			URL: fmt.Sprintf(
				"https://api.twitch.tv/helix/moderation/shield_mode?broadcaster_id=%s&moderator_id=%s",
				channelID, botID,
			),
		}),
		"executing update request",
	)
}
