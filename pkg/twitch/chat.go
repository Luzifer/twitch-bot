package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

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
