package twitch

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/mitchellh/hashstructure/v2"
)

type (
	// GetVideoOpts contain the query parameter for the GetVideos query
	//
	// See https://dev.twitch.tv/docs/api/reference/#get-videos for details
	GetVideoOpts struct {
		ID       string             // Required: Exactly one of ID, UserID, GameID
		UserID   string             // Required: Exactly one of ID, UserID, GameID
		GameID   string             // Required: Exactly one of ID, UserID, GameID
		Language string             // Optional: Use only with GameID
		Period   GetVideoOptsPeriod // Optional: Use only with GameID or UserID
		Sort     GetVideoOptsSort   // Optional: Use only with GameID or UserID
		Type     GetVideoOptsType   // Optional: Use only with GameID or UserID
		First    int64              // Optional: Use only with GameID or UserID
		After    string             // Optional: Use only with UserID
		Before   string             // Optional: Use only with UserID
	}

	// GetVideoOptsPeriod represents a filter used to filter the list of
	// videos by when they were published
	GetVideoOptsPeriod string
	// GetVideoOptsSort represents the order to sort the returned videos in
	GetVideoOptsSort string
	// GetVideoOptsType represents a filter used to filter the list of
	// videos by the video's type
	GetVideoOptsType string

	// Video contains information about a published video
	Video struct {
		ID            string    `json:"id"`
		StreamID      *string   `json:"stream_id"`
		UserID        string    `json:"user_id"`
		UserLogin     string    `json:"user_login"`
		UserName      string    `json:"user_name"`
		Title         string    `json:"title"`
		Description   string    `json:"description"`
		CreatedAt     time.Time `json:"created_at"`
		PublishedAt   time.Time `json:"published_at"`
		URL           string    `json:"url"`
		ThumbnailURL  string    `json:"thumbnail_url"`
		Viewable      string    `json:"viewable"`
		ViewCount     int64     `json:"view_count"`
		Language      string    `json:"language"`
		Type          string    `json:"type"`
		Duration      string    `json:"duration"`
		MutedSegments []struct {
			Duration int64 `json:"duration"`
			Offset   int64 `json:"offset"`
		} `json:"muted_segments"`
	}
)

// List of filters for GetVideoOpts.Period
const (
	GetVideoOptsPeriodAll   GetVideoOptsPeriod = "all"
	GetVideoOptsPeriodDay   GetVideoOptsPeriod = "day"
	GetVideoOptsPeriodMonth GetVideoOptsPeriod = "month"
	GetVideoOptsPeriodWeek  GetVideoOptsPeriod = "week"
)

// List of sort options for GetVideoOpts.Sort
const (
	GetVideoOptsSortTime     GetVideoOptsSort = "time"
	GetVideoOptsSortTrending GetVideoOptsSort = "trending"
	GetVideoOptsSortViews    GetVideoOptsSort = "views"
)

// List of types for GetVideoOpts.Type
const (
	GetVideoOptsTypeAll       GetVideoOptsType = "all"
	GetVideoOptsTypeArchive   GetVideoOptsType = "archive"
	GetVideoOptsTypeHighlight GetVideoOptsType = "highlight"
	GetVideoOptsTypeUpload    GetVideoOptsType = "upload"
)

// GetVideos fetches information about one or more published videos
func (c *Client) GetVideos(ctx context.Context, opts GetVideoOpts) (videos []Video, err error) {
	optsCacheKey, err := opts.cacheKey()
	if err != nil {
		return nil, fmt.Errorf("getting opts cache key: %w", err)
	}

	cacheKey := []string{"currentVideos", optsCacheKey}
	if vids := c.apiCache.Get(cacheKey); vids != nil {
		return vids.([]Video), nil
	}

	var payload struct {
		Data []Video `json:"data"`
	}

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeAppAccessToken,
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/videos?%s", opts.queryParams()),
	}); err != nil {
		return nil, fmt.Errorf("requesting videos: %w", err)
	}

	// Videos can be changed at any moment, cache for a short period of time
	c.apiCache.Set(cacheKey, twitchMinCacheTime, payload.Data)

	return payload.Data, nil
}

func (g GetVideoOpts) cacheKey() (string, error) {
	h, err := hashstructure.Hash(g, hashstructure.FormatV2, nil)
	if err != nil {
		return "", fmt.Errorf("hashing opts: %w", err)
	}

	return strconv.FormatUint(h, 10), nil
}

func (g GetVideoOpts) queryParams() string {
	params := url.Values{}

	for k, v := range map[string]string{
		"id":       g.ID,
		"user_id":  g.UserID,
		"game_id":  g.GameID,
		"language": g.Language,
		"period":   string(g.Period),
		"sort":     string(g.Sort),
		"type":     string(g.Type),
		"first":    strconv.FormatInt(g.First, 10),
		"after":    g.After,
		"before":   g.Before,
	} {
		if v != "" && v != "0" {
			params.Set(k, v)
		}
	}

	return params.Encode()
}
