package twitch

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
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

	broadcaster, err := c.GetIDForUsername(broadcasterName)
	if err != nil {
		return 0, 0, errors.Wrap(err, "getting ID for broadcaster name")
	}

	var data subInfo

	if err = c.Request(ClientRequestOpts{
		AuthType: AuthTypeBearerToken,
		Context:  ctx,
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &data,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/subscriptions?broadcaster_id=%s", broadcaster),
	}); err != nil {
		return 0, 0, errors.Wrap(err, "executing request")
	}

	// Lets not annoy the API but only ask every 5m
	c.apiCache.Set(cacheKey, subInfoCacheTimeout, data)

	return data.Total, data.Points, nil
}
