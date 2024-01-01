package twitch

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const pollCacheTimeout = 10 * time.Second // Cache polls for a short moment to prevent multiple requests in one template

type (
	// PollInfo contains information about a Twitch poll
	PollInfo struct {
		ID               string `json:"id"`
		BroadcasterID    string `json:"broadcaster_id"`
		BroadcasterName  string `json:"broadcaster_name"`
		BroadcasterLogin string `json:"broadcaster_login"`
		Title            string `json:"title"`
		Choices          []struct {
			ID                 string `json:"id"`
			Title              string `json:"title"`
			Votes              int    `json:"votes"`
			ChannelPointsVotes int    `json:"channel_points_votes"`
		} `json:"choices"`
		ChannelPointsVotingEnabled bool       `json:"channel_points_voting_enabled"`
		ChannelPointsPerVote       int        `json:"channel_points_per_vote"`
		Status                     string     `json:"status"`
		Duration                   int        `json:"duration"`
		StartedAt                  time.Time  `json:"started_at"`
		EndedAt                    *time.Time `json:"ended_at"`
	}
)

// GetLatestPoll returns the lastest (active or past) poll inside the
// given channel
func (c *Client) GetLatestPoll(ctx context.Context, channel string) (*PollInfo, error) {
	cacheKey := []string{"getLatestPoll", channel}
	if poll := c.apiCache.Get(cacheKey); poll != nil {
		return poll.(*PollInfo), nil
	}

	id, err := c.GetIDForUsername(ctx, channel)
	if err != nil {
		return nil, errors.Wrap(err, "getting ID for username")
	}

	var payload struct {
		Data []*PollInfo `json:"data"`
	}

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeBearerToken,
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/polls?broadcaster_id=%s&first=1", id),
	}); err != nil {
		return nil, errors.Wrap(err, "request channel info")
	}

	if len(payload.Data) < 1 {
		return nil, errors.New("no polls found")
	}

	c.apiCache.Set(cacheKey, pollCacheTimeout, payload.Data[0])

	return payload.Data[0], nil
}
