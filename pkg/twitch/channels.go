package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func (c *Client) ModifyChannelInformation(ctx context.Context, broadcasterName string, game, title *string) error {
	if game == nil && title == nil {
		return errors.New("netiher game nor title provided")
	}

	broadcaster, err := c.GetIDForUsername(broadcasterName)
	if err != nil {
		return errors.Wrap(err, "getting ID for broadcaster name")
	}

	data := struct {
		GameID *string `json:"game_id,omitempty"`
		Title  *string `json:"title,omitempty"`
	}{
		Title: title,
	}

	switch {
	case game == nil:
		// We don't set the GameID

	case (*game)[0] == '@':
		// We got an ID and don't need to resolve
		gameID := (*game)[1:]
		data.GameID = &gameID

	default:
		categories, err := c.SearchCategories(ctx, *game)
		if err != nil {
			return errors.Wrap(err, "searching for game")
		}

		switch len(categories) {
		case 0:
			return errors.New("no matching game found")

		case 1:
			data.GameID = &categories[0].ID

		default:
			// Multiple matches: Search for exact one
			for _, c := range categories {
				if c.Name == *game {
					gid := c.ID
					data.GameID = &gid
					break
				}
			}

			if data.GameID == nil {
				// No exact match found: This is an error
				return errors.New("no exact game match found")
			}
		}
	}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(data); err != nil {
		return errors.Wrap(err, "encoding payload")
	}

	return errors.Wrap(
		c.request(clientRequestOpts{
			AuthType: authTypeBearerToken,
			Body:     body,
			Context:  ctx,
			Method:   http.MethodPatch,
			OKStatus: http.StatusNoContent,
			URL:      fmt.Sprintf("https://api.twitch.tv/helix/channels?broadcaster_id=%s", broadcaster),
		}),
		"executing request",
	)
}
