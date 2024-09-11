package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type (
	// StreamInfo contains all the information known about a stream
	StreamInfo struct {
		ID           string    `json:"id"`
		UserID       string    `json:"user_id"`
		UserLogin    string    `json:"user_login"`
		UserName     string    `json:"user_name"`
		GameID       string    `json:"game_id"`
		GameName     string    `json:"game_name"`
		Type         string    `json:"type"`
		Title        string    `json:"title"`
		ViewerCount  int64     `json:"viewer_count"`
		StartedAt    time.Time `json:"started_at"`
		Language     string    `json:"language"`
		ThumbnailURL string    `json:"thumbnail_url"`
		TagIds       []string  `json:"tag_ids"` //revive:disable-line:var-naming // Disabled to prevent breaking change
		IsMature     bool      `json:"is_mature"`
	}

	// StreamMarkerInfo contains information about a marker on a stream
	StreamMarkerInfo struct {
		ID              int64     `json:"id"`
		CreatedAt       time.Time `json:"created_at"`
		Description     string    `json:"description"`
		PositionSeconds int64     `json:"position_seconds"`
	}
)

// ErrNoStreamsFound allows to differntiate between an HTTP error and
// the fact there just is no stream found
var ErrNoStreamsFound = errors.New("no streams found")

// CreateStreamMarker creates a marker for the currently running stream.
// The stream must be live, no VoD, no upload and no re-run.
// The description may be up to 140 chars and can be omitted.
func (c *Client) CreateStreamMarker(ctx context.Context, description string) (marker StreamMarkerInfo, err error) {
	body := new(bytes.Buffer)

	userID, _, err := c.GetAuthorizedUser(ctx)
	if err != nil {
		return marker, fmt.Errorf("getting ID for current user: %w", err)
	}

	if err = json.NewEncoder(body).Encode(struct {
		UserID      string `json:"user_id"`
		Description string `json:"description,omitempty"`
	}{
		UserID:      userID,
		Description: description,
	}); err != nil {
		return marker, fmt.Errorf("encoding payload: %w", err)
	}

	var payload struct {
		Data []StreamMarkerInfo `json:"data"`
	}

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeBearerToken,
		Body:     body,
		Method:   http.MethodPost,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      "https://api.twitch.tv/helix/streams/markers",
	}); err != nil {
		return marker, fmt.Errorf("creating marker: %w", err)
	}

	return payload.Data[0], nil
}

// GetCurrentStreamInfo returns the StreamInfo of the currently running
// stream of the given username
func (c *Client) GetCurrentStreamInfo(ctx context.Context, username string) (*StreamInfo, error) {
	cacheKey := []string{"currentStreamInfo", username}
	if si := c.apiCache.Get(cacheKey); si != nil {
		return si.(*StreamInfo), nil
	}

	id, err := c.GetIDForUsername(ctx, username)
	if err != nil {
		return nil, errors.Wrap(err, "getting ID for username")
	}

	var payload struct {
		Data []*StreamInfo `json:"data"`
	}

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeAppAccessToken,
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/streams?user_id=%s", id),
	}); err != nil {
		return nil, errors.Wrap(err, "request channel info")
	}

	switch l := len(payload.Data); l {
	case 0:
		return nil, ErrNoStreamsFound

	case 1:
		// That's expected

	default:
		return nil, errors.Errorf("unexpected number of streams returned: %d", l)
	}

	// Stream-info can be changed at any moment, cache for a short period of time
	c.apiCache.Set(cacheKey, twitchMinCacheTime, payload.Data[0])

	return payload.Data[0], nil
}

// GetRecentStreamInfo returns the category and the title the given
// username has configured for their recent (or next) stream
func (c *Client) GetRecentStreamInfo(ctx context.Context, username string) (category string, title string, err error) {
	cacheKey := []string{"recentStreamInfo", username}
	if d := c.apiCache.Get(cacheKey); d != nil {
		return d.([2]string)[0], d.([2]string)[1], nil
	}

	id, err := c.GetIDForUsername(ctx, username)
	if err != nil {
		return "", "", errors.Wrap(err, "getting ID for username")
	}

	var payload struct {
		Data []struct {
			BroadcasterID string `json:"broadcaster_id"`
			GameID        string `json:"game_id"`
			GameName      string `json:"game_name"`
			Title         string `json:"title"`
		} `json:"data"`
	}

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeAppAccessToken,
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/channels?broadcaster_id=%s", id),
	}); err != nil {
		return "", "", errors.Wrap(err, "request channel info")
	}

	if l := len(payload.Data); l != 1 {
		return "", "", errors.Errorf("unexpected number of users returned: %d", l)
	}

	// Stream-info can be changed at any moment, cache for a short period of time
	c.apiCache.Set(cacheKey, twitchMinCacheTime, [2]string{payload.Data[0].GameName, payload.Data[0].Title})

	return payload.Data[0].GameName, payload.Data[0].Title, nil
}

// HasLiveStream checks whether the given user is currently streaming
func (c *Client) HasLiveStream(ctx context.Context, username string) (bool, error) {
	cacheKey := []string{"hasLiveStream", username}
	if d := c.apiCache.Get(cacheKey); d != nil {
		return d.(bool), nil
	}

	var payload struct {
		Data []struct {
			ID        string `json:"id"`
			UserLogin string `json:"user_login"`
			Type      string `json:"type"`
		} `json:"data"`
	}

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeAppAccessToken,
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/streams?user_login=%s", username),
	}); err != nil {
		return false, errors.Wrap(err, "request stream info")
	}

	// Live status might change recently, cache for one minute
	c.apiCache.Set(cacheKey, twitchMinCacheTime, len(payload.Data) == 1 && payload.Data[0].Type == "live")

	return len(payload.Data) == 1 && payload.Data[0].Type == "live", nil
}
