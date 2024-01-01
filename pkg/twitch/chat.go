package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
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
		return errors.Wrap(err, "getting bot user-id")
	}

	channelID, err := c.GetIDForUsername(ctx, strings.TrimLeft(channel, "#@"))
	if err != nil {
		return errors.Wrap(err, "getting channel user-id")
	}

	body := new(bytes.Buffer)
	if err = json.NewEncoder(body).Encode(payload); err != nil {
		return errors.Wrap(err, "encoding payload")
	}

	return errors.Wrap(
		c.Request(ctx, ClientRequestOpts{
			AuthType: AuthTypeBearerToken,
			Method:   http.MethodPost,
			OKStatus: http.StatusNoContent,
			Body:     body,
			URL: fmt.Sprintf(
				"https://api.twitch.tv/helix/chat/announcements?broadcaster_id=%s&moderator_id=%s",
				channelID, botID,
			),
		}),
		"executing request",
	)
}

// SendShoutout creates a Twitch-native shoutout in the given channel
// for the given user. This equals `/shoutout <user>` in the channel.
func (c *Client) SendShoutout(ctx context.Context, channel, user string) error {
	botID, _, err := c.GetAuthorizedUser(ctx)
	if err != nil {
		return errors.Wrap(err, "getting bot user-id")
	}

	channelID, err := c.GetIDForUsername(ctx, strings.TrimLeft(channel, "#@"))
	if err != nil {
		return errors.Wrap(err, "getting channel user-id")
	}

	userID, err := c.GetIDForUsername(ctx, strings.TrimLeft(user, "#@"))
	if err != nil {
		return errors.Wrap(err, "getting user user-id")
	}

	params := make(url.Values)
	params.Set("from_broadcaster_id", channelID)
	params.Set("moderator_id", botID)
	params.Set("to_broadcaster_id", userID)

	return errors.Wrap(
		c.Request(ctx, ClientRequestOpts{
			AuthType: AuthTypeBearerToken,
			Method:   http.MethodPost,
			OKStatus: http.StatusNoContent,
			URL: fmt.Sprintf(
				"https://api.twitch.tv/helix/chat/shoutouts?%s",
				params.Encode(),
			),
		}),
		"executing request",
	)
}
