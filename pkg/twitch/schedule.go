package twitch

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type (
	// ChannelStreamSchedule represents the schedule of a channels with
	// its segments represening single planned streams
	ChannelStreamSchedule struct {
		Segments         []ChannelStreamScheduleSegment `json:"segments"`
		BroadcasterID    string                         `json:"broadcaster_id"`
		BroadcasterName  string                         `json:"broadcaster_name"`
		BroadcasterLogin string                         `json:"broadcaster_login"`
		Vacation         struct {
			StartTime time.Time `json:"start_time"`
			EndTime   time.Time `json:"end_time"`
		} `json:"vacation"`
	}

	// ChannelStreamScheduleSegment represents a single stream inside the
	// ChannelStreamSchedule
	ChannelStreamScheduleSegment struct {
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
	}
)

// GetChannelStreamSchedule gets the broadcaster’s streaming schedule
func (c *Client) GetChannelStreamSchedule(ctx context.Context, channel string) (*ChannelStreamSchedule, error) {
	channelID, err := c.GetIDForUsername(ctx, strings.TrimLeft(channel, "#@"))
	if err != nil {
		return nil, fmt.Errorf("getting channel user-id: %w", err)
	}

	var payload struct {
		Data *ChannelStreamSchedule `json:"data"`
	}

	if err = c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeAppAccessToken,
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/schedule?broadcaster_id=%s", channelID),
	}); err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	return payload.Data, nil
}
