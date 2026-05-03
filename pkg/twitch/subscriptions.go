package twitch

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const subInfoCacheTimeout = 300 * time.Second

type (
	subInfo struct {
		Total  int64 `json:"total"`
		Points int64 `json:"points"`
	}
)

// GetBroadcasterSubscriptionCount gets a list of users that subscribe to the specified broadcaster.
func (c *Client) GetBroadcasterSubscriptionCount(ctx context.Context, broadcasterName string) (subCount, subPoints int64, err error) {
	cacheKey := []string{"broadcasterSubscriptionCountByChannel", broadcasterName}
	if d := c.apiCache.Get(cacheKey); d != nil {
		data := d.(subInfo)
		return data.Total, data.Points, nil
	}

	broadcaster, err := c.GetIDForUsername(ctx, broadcasterName)
	if err != nil {
		return 0, 0, fmt.Errorf("getting ID for broadcaster name: %w", err)
	}

	var data subInfo

	if err = c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeBearerToken,
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &data,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/subscriptions?broadcaster_id=%s", broadcaster),
	}); err != nil {
		return 0, 0, fmt.Errorf("executing request: %w", err)
	}

	// Lets not annoy the API but only ask every 5m
	c.apiCache.Set(cacheKey, subInfoCacheTimeout, data)

	return data.Total, data.Points, nil
}
