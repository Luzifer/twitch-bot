package twitch

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const clipCacheTimeout = 10 * time.Minute // Clips do not change that fast

type (
	// ClipInfo contains the information about a clip
	ClipInfo struct {
		ID              string    `json:"id"`
		URL             string    `json:"url"`
		EmbedURL        string    `json:"embed_url"`
		BroadcasterID   string    `json:"broadcaster_id"`
		BroadcasterName string    `json:"broadcaster_name"`
		CreatorID       string    `json:"creator_id"`
		CreatorName     string    `json:"creator_name"`
		VideoID         string    `json:"video_id"`
		GameID          string    `json:"game_id"`
		Language        string    `json:"language"`
		Title           string    `json:"title"`
		ViewCount       int64     `json:"view_count"`
		CreatedAt       time.Time `json:"created_at"`
		ThumbnailURL    string    `json:"thumbnail_url"`
		Duration        float64   `json:"duration"`
		VodOffset       int64     `json:"vod_offset"`
	}

	// CreateClipResponse contains the API response to a create clip call
	CreateClipResponse struct {
		ID      string `json:"id"`
		EditURL string `json:"edit_url"`
	}
)

// CreateClip triggers the creation of a clip in the given channel.
// If addDelay is true an artificial delay will be added (for
// broadcasters who trigger this function already knowing something
// will happen but not yet visible in stream).
func (c *Client) CreateClip(ctx context.Context, channel string, addDelay bool) (ccr CreateClipResponse, err error) {
	id, err := c.GetIDForUsername(ctx, channel)
	if err != nil {
		return ccr, errors.Wrap(err, "getting ID for channel")
	}

	var payload struct {
		Data []CreateClipResponse
	}

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeBearerToken,
		Method:   http.MethodPost,
		OKStatus: http.StatusAccepted,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/clips?broadcaster_id=%s&has_delay=%v", id, addDelay),
	}); err != nil {
		return ccr, errors.Wrap(err, "triggering clip create")
	}

	if l := len(payload.Data); l != 1 {
		return ccr, errors.Errorf("unexpected number of results returned: %d", l)
	}

	return payload.Data[0], nil
}

// GetClipByID gets a video clip that were captured from streams by
// its ID (slug in the URL)
func (c *Client) GetClipByID(ctx context.Context, clipID string) (ClipInfo, error) {
	cacheKey := []string{"getClipByID", clipID}
	if clip := c.apiCache.Get(cacheKey); clip != nil {
		return clip.(ClipInfo), nil
	}

	var payload struct {
		Data []ClipInfo
	}

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeAppAccessToken,
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/clips?id=%s", clipID),
	}); err != nil {
		return ClipInfo{}, errors.Wrap(err, "getting clip info")
	}

	if l := len(payload.Data); l != 1 {
		return ClipInfo{}, errors.Errorf("unexpected number of clip info returned: %d", l)
	}

	c.apiCache.Set(cacheKey, clipCacheTimeout, payload.Data[0])

	return payload.Data[0], nil
}
