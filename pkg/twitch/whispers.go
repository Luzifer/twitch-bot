package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SendWhisper sends a whisper from the bot to the specified user.
//
// For details about message limits see the official documentation:
// https://dev.twitch.tv/docs/api/reference#send-whisper
func (c *Client) SendWhisper(ctx context.Context, toUser, message string) error {
	var payload struct {
		Message string `json:"message"`
	}

	payload.Message = message

	botID, _, err := c.GetAuthorizedUser(ctx)
	if err != nil {
		return fmt.Errorf("getting bot user-id: %w", err)
	}

	targetID, err := c.GetIDForUsername(ctx, toUser)
	if err != nil {
		return fmt.Errorf("getting target user-id: %w", err)
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
			"https://api.twitch.tv/helix/whispers?from_user_id=%s&to_user_id=%s",
			botID, targetID,
		),
	}); err != nil {
		return fmt.Errorf("executing whisper request: %w", err)
	}

	return nil
}
