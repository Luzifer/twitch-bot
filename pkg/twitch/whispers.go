package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// SendWhisper sends a whisper from the bot to the specified user.
//
// For details about message limits see the official documentation:
// https://dev.twitch.tv/docs/api/reference#send-whisper
func (c *Client) SendWhisper(toUser, message string) error {
	var payload struct {
		Message string `json:"message"`
	}

	payload.Message = message

	botID, _, err := c.GetAuthorizedUser()
	if err != nil {
		return errors.Wrap(err, "getting bot user-id")
	}

	targetID, err := c.GetIDForUsername(toUser)
	if err != nil {
		return errors.Wrap(err, "getting target user-id")
	}

	body := new(bytes.Buffer)
	if err = json.NewEncoder(body).Encode(payload); err != nil {
		return errors.Wrap(err, "encoding payload")
	}

	return errors.Wrap(
		c.Request(ClientRequestOpts{
			AuthType: AuthTypeBearerToken,
			Context:  context.Background(),
			Method:   http.MethodPost,
			OKStatus: http.StatusNoContent,
			Body:     body,
			URL: fmt.Sprintf(
				"https://api.twitch.tv/helix/whispers?from_user_id=%s&to_user_id=%s",
				botID, targetID,
			),
		}),
		"executing whisper request",
	)
}
