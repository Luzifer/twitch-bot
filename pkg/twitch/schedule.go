package twitch

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type (
	// ChannelStreamSchedule represents the schedule of a channels with
	// its segments represening single planned streams
	ChannelStreamSchedule struct {
		Segments []struct {
			ID            string     `json:"id"`
			StartTime     time.Time  `json:"start_time"`
			EndTime       time.Time  `json:"end_time"`
			Title         string     `json:"title"`
			CanceledUntil *time.Time `json:"canceled_until"`
			Category      struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"category"`
			IsRecurring bool `json:"is_recurring"`
		} `json:"segments"`
		BroadcasterID    string `json:"broadcaster_id"`
		BroadcasterName  string `json:"broadcaster_name"`
		BroadcasterLogin string `json:"broadcaster_login"`
		Vacation         struct {
			StartTime time.Time `json:"start_time"`
			EndTime   time.Time `json:"end_time"`
		} `json:"vacation"`
	}
)

// GetChannelStreamSchedule gets the broadcaster’s streaming schedule
func (c *Client) GetChannelStreamSchedule(ctx context.Context, channel string) (*ChannelStreamSchedule, error) {
	channelID, err := c.GetIDForUsername(strings.TrimLeft(channel, "#@"))
	if err != nil {
		return nil, errors.Wrap(err, "getting channel user-id")
	}

	var payload struct {
		Data *ChannelStreamSchedule `json:"data"`
	}

	return payload.Data, errors.Wrap(
		c.Request(ClientRequestOpts{
			AuthType: AuthTypeAppAccessToken,
			Context:  ctx,
			Method:   http.MethodGet,
			OKStatus: http.StatusOK,
			Out:      &payload,
			URL:      fmt.Sprintf("https://api.twitch.tv/helix/schedule?broadcaster_id=%s", channelID),
		}),
		"executing request",
	)
}
