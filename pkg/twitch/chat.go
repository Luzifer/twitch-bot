package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// SendChatAnnouncement sends an announcement in the specified
// channel with the given message. Colors must be blue, green,
// orange, purple or primary (empty color = primary)
func (c *Client) SendChatAnnouncement(channel, color, message string) error {
	var payload struct {
		Color   string `json:"color,omitempty"`
		Message string `json:"message"`
	}

	payload.Color = color
	payload.Message = message

	botID, _, err := c.GetAuthorizedUser()
	if err != nil {
		return errors.Wrap(err, "getting bot user-id")
	}

	channelID, err := c.GetIDForUsername(channel)
	if err != nil {
		return errors.Wrap(err, "getting channel user-id")
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
