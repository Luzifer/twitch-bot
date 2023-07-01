package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

func (c *Client) AddChannelVIP(ctx context.Context, broadcasterName, userName string) error {
	broadcaster, err := c.GetIDForUsername(broadcasterName)
	if err != nil {
		return errors.Wrap(err, "getting ID for broadcaster name")
	}

	userID, err := c.GetIDForUsername(userName)
	if err != nil {
		return errors.Wrap(err, "getting ID for user name")
	}

	return errors.Wrap(
		c.Request(ClientRequestOpts{
			AuthType: AuthTypeBearerToken,
			Context:  ctx,
			Method:   http.MethodPost,
			OKStatus: http.StatusNoContent,
			URL:      fmt.Sprintf("https://api.twitch.tv/helix/channels/vips?broadcaster_id=%s&user_id=%s", broadcaster, userID),
		}),
		"executing request",
	)
}

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
				if strings.EqualFold(c.Name, *game) {
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
		c.Request(ClientRequestOpts{
			AuthType: AuthTypeBearerToken,
			Body:     body,
			Context:  ctx,
			Method:   http.MethodPatch,
			OKStatus: http.StatusNoContent,
			URL:      fmt.Sprintf("https://api.twitch.tv/helix/channels?broadcaster_id=%s", broadcaster),
		}),
		"executing request",
	)
}

func (c *Client) RemoveChannelVIP(ctx context.Context, broadcasterName, userName string) error {
	broadcaster, err := c.GetIDForUsername(broadcasterName)
	if err != nil {
		return errors.Wrap(err, "getting ID for broadcaster name")
	}

	userID, err := c.GetIDForUsername(userName)
	if err != nil {
		return errors.Wrap(err, "getting ID for user name")
	}

	return errors.Wrap(
		c.Request(ClientRequestOpts{
			AuthType: AuthTypeBearerToken,
			Context:  ctx,
			Method:   http.MethodDelete,
			OKStatus: http.StatusNoContent,
			URL:      fmt.Sprintf("https://api.twitch.tv/helix/channels/vips?broadcaster_id=%s&user_id=%s", broadcaster, userID),
		}),
		"executing request",
	)
}

// RunCommercial starts a commercial on the specified channel
func (c *Client) RunCommercial(ctx context.Context, channel string, duration int64) error {
	channelID, err := c.GetIDForUsername(channel)
	if err != nil {
		return errors.Wrap(err, "getting ID for channel name")
	}

	payload := struct {
		BroadcasterID string `json:"broadcaster_id"`
		Length        int64  `json:"length"`
	}{
		BroadcasterID: channelID,
		Length:        duration,
	}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(payload); err != nil {
		return errors.Wrap(err, "encoding payload")
	}

	return errors.Wrap(
		c.Request(ClientRequestOpts{
			AuthType: AuthTypeBearerToken,
			Body:     body,
			Context:  ctx,
			Method:   http.MethodPost,
			OKStatus: http.StatusOK,
			URL:      "https://api.twitch.tv/helix/channels/commercial",
		}),
		"executing request",
	)
}
