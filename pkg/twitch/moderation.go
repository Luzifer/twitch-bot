package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	errMessageAlreadyBanned = "The user specified in the user_id field is already banned."
	maxTimeoutDuration      = 1209600 * time.Second
)

// BanUser bans or timeouts a user in the given channel. Setting the
// duration to 0 will result in a ban, setting if greater than 0 will
// result in a timeout. The timeout is automatically converted to
// full seconds. The timeout duration must be less than 1209600s.
func (c *Client) BanUser(ctx context.Context, channel, username string, duration time.Duration, reason string) error {
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

	botID, _, err := c.GetAuthorizedUser(ctx)
	if err != nil {
		return errors.Wrap(err, "getting bot user-id")
	}

	channelID, err := c.GetIDForUsername(ctx, strings.TrimLeft(channel, "#@"))
	if err != nil {
		return errors.Wrap(err, "getting channel user-id")
	}

	if payload.Data.UserID, err = c.GetIDForUsername(ctx, username); err != nil {
		return errors.Wrap(err, "getting target user-id")
	}

	body := new(bytes.Buffer)
	if err = json.NewEncoder(body).Encode(payload); err != nil {
		return errors.Wrap(err, "encoding payload")
	}

	return errors.Wrapf(
		c.Request(ctx, ClientRequestOpts{
			AuthType: AuthTypeBearerToken,
			Method:   http.MethodPost,
			OKStatus: http.StatusOK,
			Body:     body,
			URL: fmt.Sprintf(
				"https://api.twitch.tv/helix/moderation/bans?broadcaster_id=%s&moderator_id=%s",
				channelID, botID,
			),
			ValidateFunc: func(opts ClientRequestOpts, resp *http.Response) error {
				if resp.StatusCode == http.StatusBadRequest {
					// The user might already be banned, lets check the error in detail
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						return newHTTPError(resp.StatusCode, nil, err)
					}

					var payload ErrorResponse
					if err = json.Unmarshal(body, &payload); err == nil && payload.Message == errMessageAlreadyBanned {
						// The user is already banned, that's fine as that was
						// our goal!
						return nil
					}

					return newHTTPError(resp.StatusCode, body, nil)
				}

				return ValidateStatus(opts, resp)
			},
		}),
		"executing ban request for %q in %q", username, channel,
	)
}

// DeleteMessage deletes one or all messages from the specified chat.
// If no messageID is given all messages are deleted. If a message ID
// is given the message must be no older than 6 hours and it must not
// be posted by broadcaster or moderator.
func (c *Client) DeleteMessage(ctx context.Context, channel, messageID string) error {
	botID, _, err := c.GetAuthorizedUser(ctx)
	if err != nil {
		return errors.Wrap(err, "getting bot user-id")
	}

	channelID, err := c.GetIDForUsername(ctx, strings.TrimLeft(channel, "#@"))
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
		c.Request(ctx, ClientRequestOpts{
			AuthType: AuthTypeBearerToken,
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
func (c *Client) UnbanUser(ctx context.Context, channel, username string) error {
	botID, _, err := c.GetAuthorizedUser(ctx)
	if err != nil {
		return errors.Wrap(err, "getting bot user-id")
	}

	channelID, err := c.GetIDForUsername(ctx, strings.TrimLeft(channel, "#@"))
	if err != nil {
		return errors.Wrap(err, "getting channel user-id")
	}

	userID, err := c.GetIDForUsername(ctx, username)
	if err != nil {
		return errors.Wrap(err, "getting target user-id")
	}

	return errors.Wrap(
		c.Request(ctx, ClientRequestOpts{
			AuthType: AuthTypeBearerToken,
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
	botID, _, err := c.GetAuthorizedUser(ctx)
	if err != nil {
		return errors.Wrap(err, "getting bot user-id")
	}

	channelID, err := c.GetIDForUsername(ctx, strings.TrimLeft(channel, "#@"))
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
		c.Request(ctx, ClientRequestOpts{
			AuthType: AuthTypeBearerToken,
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
