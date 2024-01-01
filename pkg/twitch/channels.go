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

// AddChannelVIP adds the given user as a VIP in the given channel
func (c *Client) AddChannelVIP(ctx context.Context, channel, userName string) error {
	broadcaster, err := c.GetIDForUsername(ctx, channel)
	if err != nil {
		return errors.Wrap(err, "getting ID for channel name")
	}

	userID, err := c.GetIDForUsername(ctx, userName)
	if err != nil {
		return errors.Wrap(err, "getting ID for user name")
	}

	return errors.Wrap(
		c.Request(ctx, ClientRequestOpts{
			AuthType: AuthTypeBearerToken,
			Method:   http.MethodPost,
			OKStatus: http.StatusNoContent,
			URL:      fmt.Sprintf("https://api.twitch.tv/helix/channels/vips?broadcaster_id=%s&user_id=%s", broadcaster, userID),
		}),
		"executing request",
	)
}

// ModifyChannelInformation adjusts category and title for the given
// channel
func (c *Client) ModifyChannelInformation(ctx context.Context, channel string, category, title *string) error {
	if category == nil && title == nil {
		return errors.New("netiher game nor title provided")
	}

	broadcaster, err := c.GetIDForUsername(ctx, channel)
	if err != nil {
		return errors.Wrap(err, "getting ID for channel name")
	}

	data := struct {
		GameID *string `json:"game_id,omitempty"`
		Title  *string `json:"title,omitempty"`
	}{
		Title: title,
	}

	switch {
	case category == nil:
		// We don't set the GameID

	case (*category)[0] == '@':
		// We got an ID and don't need to resolve
		gameID := (*category)[1:]
		data.GameID = &gameID

	default:
		categories, err := c.SearchCategories(ctx, *category)
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
				if strings.EqualFold(c.Name, *category) {
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
		c.Request(ctx, ClientRequestOpts{
			AuthType: AuthTypeBearerToken,
			Body:     body,
			Method:   http.MethodPatch,
			OKStatus: http.StatusNoContent,
			URL:      fmt.Sprintf("https://api.twitch.tv/helix/channels?broadcaster_id=%s", broadcaster),
		}),
		"executing request",
	)
}

// RemoveChannelVIP removes the given user as a VIP in the given channel
func (c *Client) RemoveChannelVIP(ctx context.Context, channel, userName string) error {
	broadcaster, err := c.GetIDForUsername(ctx, channel)
	if err != nil {
		return errors.Wrap(err, "getting ID for channel name")
	}

	userID, err := c.GetIDForUsername(ctx, userName)
	if err != nil {
		return errors.Wrap(err, "getting ID for user name")
	}

	return errors.Wrap(
		c.Request(ctx, ClientRequestOpts{
			AuthType: AuthTypeBearerToken,
			Method:   http.MethodDelete,
			OKStatus: http.StatusNoContent,
			URL:      fmt.Sprintf("https://api.twitch.tv/helix/channels/vips?broadcaster_id=%s&user_id=%s", broadcaster, userID),
		}),
		"executing request",
	)
}

// RunCommercial starts a commercial on the specified channel
func (c *Client) RunCommercial(ctx context.Context, channel string, duration int64) error {
	channelID, err := c.GetIDForUsername(ctx, channel)
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
		c.Request(ctx, ClientRequestOpts{
			AuthType: AuthTypeBearerToken,
			Body:     body,
			Method:   http.MethodPost,
			OKStatus: http.StatusOK,
			URL:      "https://api.twitch.tv/helix/channels/commercial",
		}),
		"executing request",
	)
}
