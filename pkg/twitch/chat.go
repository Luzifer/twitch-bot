package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	pinnedMessageMinDuration = 30 * time.Second
	pinnedMessageMaxDuration = 1800 * time.Second
)

type (
	// PinnedChatMessage represents a pinned message in a chat
	PinnedChatMessage struct {
		MessageID         string `json:"message_id"`
		BroadcasterID     string `json:"broadcaster_id"`
		SenderUserID      string `json:"sender_user_id"`
		SenderUserLogin   string `json:"sender_user_login"`
		SenderUserName    string `json:"sender_user_name"`
		PinnedByUserID    string `json:"pinned_by_user_id"`
		PinnedByUserLogin string `json:"pinned_by_user_login"`
		PinnedByUserName  string `json:"pinned_by_user_name"`
		Message           struct {
			Text      string `json:"text"`
			Fragments []struct {
				Type      string `json:"type"`
				Text      string `json:"text"`
				Cheermote *struct {
					Prefix string `json:"prefix"`
					Bits   int64  `json:"bits"`
					Tier   int    `json:"tier"`
				} `json:"cheermote"`
				Emote *struct {
					ID         string   `json:"id"`
					EmoteSetID string   `json:"emote_set_id"`
					OwnerID    string   `json:"owner_id"`
					Format     []string `json:"format"`
				} `json:"emote"`
				Mention *struct {
					UserID    string `json:"user_id"`
					UserLogin string `json:"user_login"`
					UserName  string `json:"user_name"`
				} `json:"mention"`
			} `json:"fragments"`
		} `json:"message"`
		StartsAt  time.Time  `json:"starts_at"`
		EndsAt    *time.Time `json:"ends_at"`
		UpdatedAt time.Time  `json:"updated_at"`
	}
)

// ErrNoPinnedChatMessage signals there was no pinned message returned
// by Twitch API for the GetPinnedChatMessage request
var ErrNoPinnedChatMessage = errors.New("no message is pinned")

// GetPinnedChatMessage gets the currently pinned message for the
// specified broadcaster’s chat room, including message fragments.
func (c *Client) GetPinnedChatMessage(ctx context.Context, channel string) (msg PinnedChatMessage, err error) {
	var payload struct {
		Data []PinnedChatMessage `json:"data"`
	}

	botID, _, err := c.GetAuthorizedUser(ctx)
	if err != nil {
		return msg, fmt.Errorf("getting bot user-id: %w", err)
	}

	channelID, err := c.GetIDForUsername(ctx, strings.TrimLeft(channel, "#@"))
	if err != nil {
		return msg, fmt.Errorf("getting channel user-id: %w", err)
	}

	params := make(url.Values)
	params.Set("broadcaster_id", channelID)
	params.Set("moderator_id", botID)

	if err = c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeBearerToken,
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/chat/pins?%s", params.Encode()),
	}); err != nil {
		return msg, fmt.Errorf("executing request: %w", err)
	}

	if len(payload.Data) == 0 {
		return msg, ErrNoPinnedChatMessage
	}

	return payload.Data[0], nil
}

// PinChatMessage pins a chat message to the top of the specified
// broadcaster’s chat room. Only one mod-pinned message can be active
// per channel at a time. If a mod-pinned message already exists, it
// is automatically replaced.
//
// duration must be between 30 and 1800 seconds and will be truncated
// to full seconds; if set to 0 message is pinned until end-of-stream
func (c *Client) PinChatMessage(ctx context.Context, channel, messageID string, duration time.Duration) (err error) {
	if duration > 0 && (duration < pinnedMessageMinDuration || duration > pinnedMessageMaxDuration) {
		return fmt.Errorf("duration must be between 30 and 1800 seconds, is %s", duration)
	}

	botID, _, err := c.GetAuthorizedUser(ctx)
	if err != nil {
		return fmt.Errorf("getting bot user-id: %w", err)
	}

	channelID, err := c.GetIDForUsername(ctx, strings.TrimLeft(channel, "#@"))
	if err != nil {
		return fmt.Errorf("getting channel user-id: %w", err)
	}

	params := make(url.Values)
	params.Set("broadcaster_id", channelID)
	params.Set("moderator_id", botID)
	params.Set("message_id", messageID)

	if duration > 0 {
		params.Set("duration_seconds", strconv.FormatInt(int64(duration.Truncate(time.Second)/time.Second), 10))
	}

	if err = c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeBearerToken,
		Method:   http.MethodPut,
		OKStatus: http.StatusNoContent,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/chat/pins?%s", params.Encode()),
	}); err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	return nil
}

// SendChatAnnouncement sends an announcement in the specified
// channel with the given message. Colors must be blue, green,
// orange, purple or primary (empty color = primary)
func (c *Client) SendChatAnnouncement(ctx context.Context, channel, color, message string) error {
	var payload struct {
		Color   string `json:"color,omitempty"`
		Message string `json:"message"`
	}

	payload.Color = color
	payload.Message = message

	botID, _, err := c.GetAuthorizedUser(ctx)
	if err != nil {
		return fmt.Errorf("getting bot user-id: %w", err)
	}

	channelID, err := c.GetIDForUsername(ctx, strings.TrimLeft(channel, "#@"))
	if err != nil {
		return fmt.Errorf("getting channel user-id: %w", err)
	}

	body := new(bytes.Buffer)
	if err = json.NewEncoder(body).Encode(payload); err != nil {
		return fmt.Errorf("encoding payload: %w", err)
	}

	if err = c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeBearerToken,
		Method:   http.MethodPost,
		OKStatus: http.StatusNoContent,
		Body:     body,
		URL: fmt.Sprintf(
			"https://api.twitch.tv/helix/chat/announcements?broadcaster_id=%s&moderator_id=%s",
			channelID, botID,
		),
	}); err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	return nil
}

// SendShoutout creates a Twitch-native shoutout in the given channel
// for the given user. This equals `/shoutout <user>` in the channel.
func (c *Client) SendShoutout(ctx context.Context, channel, user string) error {
	botID, _, err := c.GetAuthorizedUser(ctx)
	if err != nil {
		return fmt.Errorf("getting bot user-id: %w", err)
	}

	channelID, err := c.GetIDForUsername(ctx, strings.TrimLeft(channel, "#@"))
	if err != nil {
		return fmt.Errorf("getting channel user-id: %w", err)
	}

	userID, err := c.GetIDForUsername(ctx, strings.TrimLeft(user, "#@"))
	if err != nil {
		return fmt.Errorf("getting user user-id: %w", err)
	}

	params := make(url.Values)
	params.Set("from_broadcaster_id", channelID)
	params.Set("moderator_id", botID)
	params.Set("to_broadcaster_id", userID)

	if err = c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeBearerToken,
		Method:   http.MethodPost,
		OKStatus: http.StatusNoContent,
		URL: fmt.Sprintf(
			"https://api.twitch.tv/helix/chat/shoutouts?%s",
			params.Encode(),
		),
	}); err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	return nil
}

// UnpinChatMessage unpins a pinned chat message from the specified
// broadcaster’s chat room.
func (c *Client) UnpinChatMessage(ctx context.Context, channel, messageID string) (err error) {
	botID, _, err := c.GetAuthorizedUser(ctx)
	if err != nil {
		return fmt.Errorf("getting bot user-id: %w", err)
	}

	channelID, err := c.GetIDForUsername(ctx, strings.TrimLeft(channel, "#@"))
	if err != nil {
		return fmt.Errorf("getting channel user-id: %w", err)
	}

	params := make(url.Values)
	params.Set("broadcaster_id", channelID)
	params.Set("moderator_id", botID)
	params.Set("message_id", messageID)

	if err = c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeBearerToken,
		Method:   http.MethodDelete,
		OKStatus: http.StatusNoContent,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/chat/pins?%s", params.Encode()),
	}); err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	return nil
}

// UpdatePinnedChatMessage updates the duration of an existing pinned
// chat message.
//
// duration must be between 30 and 1800 seconds and will be truncated
// to full seconds; if set to 0 message is pinned until end-of-stream
func (c *Client) UpdatePinnedChatMessage(ctx context.Context, channel, messageID string, duration time.Duration) (err error) {
	if duration > 0 && (duration < pinnedMessageMinDuration || duration > pinnedMessageMaxDuration) {
		return fmt.Errorf("duration must be between 30 and 1800 seconds, is %s", duration)
	}

	botID, _, err := c.GetAuthorizedUser(ctx)
	if err != nil {
		return fmt.Errorf("getting bot user-id: %w", err)
	}

	channelID, err := c.GetIDForUsername(ctx, strings.TrimLeft(channel, "#@"))
	if err != nil {
		return fmt.Errorf("getting channel user-id: %w", err)
	}

	params := make(url.Values)
	params.Set("broadcaster_id", channelID)
	params.Set("moderator_id", botID)
	params.Set("message_id", messageID)

	if duration > 0 {
		params.Set("duration_seconds", strconv.FormatInt(int64(duration.Truncate(time.Second)/time.Second), 10))
	}

	if err = c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeBearerToken,
		Method:   http.MethodPatch,
		OKStatus: http.StatusNoContent,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/chat/pins?%s", params.Encode()),
	}); err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	return nil
}
