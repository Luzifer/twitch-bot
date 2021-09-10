package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Luzifer/go_helpers/v2/backoff"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	timeDay = 24 * time.Hour

	twitchMinCacheTime = time.Second * 30

	twitchRequestRetries = 5
	twitchRequestTimeout = 2 * time.Second
)

type (
	Category struct {
		BoxArtURL string `json:"box_art_url"`
		ID        string `json:"id"`
		Name      string `json:"name"`
	}

	Client struct {
		clientID string
		token    string

		apiCache *APICache
	}
)

func New(clientID, token string) *Client {
	return &Client{
		clientID: clientID,
		token:    token,

		apiCache: newTwitchAPICache(),
	}
}

func (c Client) APICache() *APICache { return c.apiCache }

func (c Client) GetAuthorizedUsername() (string, error) {
	var payload struct {
		Data []struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"data"`
	}

	if err := c.request(
		context.Background(),
		http.MethodGet,
		"https://api.twitch.tv/helix/users",
		nil,
		&payload,
	); err != nil {
		return "", errors.Wrap(err, "request channel info")
	}

	if l := len(payload.Data); l != 1 {
		return "", errors.Errorf("unexpected number of users returned: %d", l)
	}

	return payload.Data[0].Login, nil
}

func (c Client) GetDisplayNameForUser(username string) (string, error) {
	cacheKey := []string{"displayNameForUsername", username}
	if d := c.apiCache.Get(cacheKey); d != nil {
		return d.(string), nil
	}

	var payload struct {
		Data []struct {
			ID          string `json:"id"`
			DisplayName string `json:"display_name"`
			Login       string `json:"login"`
		} `json:"data"`
	}

	if err := c.request(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf("https://api.twitch.tv/helix/users?login=%s", username),
		nil,
		&payload,
	); err != nil {
		return "", errors.Wrap(err, "request channel info")
	}

	if l := len(payload.Data); l != 1 {
		return "", errors.Errorf("unexpected number of users returned: %d", l)
	}

	// The DisplayName for an username will not change (often), cache for a decent time
	c.apiCache.Set(cacheKey, time.Hour, payload.Data[0].DisplayName)

	return payload.Data[0].DisplayName, nil
}

func (c Client) GetFollowDate(from, to string) (time.Time, error) {
	cacheKey := []string{"followDate", from, to}
	if d := c.apiCache.Get(cacheKey); d != nil {
		return d.(time.Time), nil
	}

	fromID, err := c.GetIDForUsername(from)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "getting id for 'from' user")
	}
	toID, err := c.GetIDForUsername(to)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "getting id for 'to' user")
	}

	var payload struct {
		Data []struct {
			FollowedAt time.Time `json:"followed_at"`
		} `json:"data"`
	}

	if err := c.request(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf("https://api.twitch.tv/helix/users/follows?to_id=%s&from_id=%s", toID, fromID),
		nil,
		&payload,
	); err != nil {
		return time.Time{}, errors.Wrap(err, "request follow info")
	}

	if l := len(payload.Data); l != 1 {
		return time.Time{}, errors.Errorf("unexpected number of records returned: %d", l)
	}

	// Follow date will not change that often, cache for a long time
	c.apiCache.Set(cacheKey, timeDay, payload.Data[0].FollowedAt)

	return payload.Data[0].FollowedAt, nil
}

func (c Client) SearchCategories(ctx context.Context, name string) ([]Category, error) {
	var out []Category

	params := make(url.Values)
	params.Set("query", name)
	params.Set("first", "100")

	var resp struct {
		Data       []Category `json:"data"`
		Pagination struct {
			Cursor string `json:"cursor"`
		} `json:"pagination"`
	}

	for {
		if err := c.request(ctx, http.MethodGet, fmt.Sprintf("https://api.twitch.tv/helix/search/categories?%s", params.Encode()), nil, &resp); err != nil {
			return nil, errors.Wrap(err, "executing request")
		}

		out = append(out, resp.Data...)

		if resp.Pagination.Cursor == "" {
			break
		}

		params.Set("after", resp.Pagination.Cursor)
		resp.Pagination.Cursor = "" // Clear from struct as struct is reused
	}

	return out, nil
}

func (c Client) HasLiveStream(username string) (bool, error) {
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

	if err := c.request(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf("https://api.twitch.tv/helix/streams?user_login=%s", username),
		nil,
		&payload,
	); err != nil {
		return false, errors.Wrap(err, "request stream info")
	}

	// Live status might change recently, cache for one minute
	c.apiCache.Set(cacheKey, twitchMinCacheTime, len(payload.Data) == 1 && payload.Data[0].Type == "live")

	return len(payload.Data) == 1 && payload.Data[0].Type == "live", nil
}

func (c Client) GetIDForUsername(username string) (string, error) {
	cacheKey := []string{"idForUsername", username}
	if d := c.apiCache.Get(cacheKey); d != nil {
		return d.(string), nil
	}

	var payload struct {
		Data []struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"data"`
	}

	if err := c.request(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf("https://api.twitch.tv/helix/users?login=%s", username),
		nil,
		&payload,
	); err != nil {
		return "", errors.Wrap(err, "request channel info")
	}

	if l := len(payload.Data); l != 1 {
		return "", errors.Errorf("unexpected number of users returned: %d", l)
	}

	// The ID for an username will not change (often), cache for a long time
	c.apiCache.Set(cacheKey, timeDay, payload.Data[0].ID)

	return payload.Data[0].ID, nil
}

func (c Client) GetRecentStreamInfo(username string) (string, string, error) {
	cacheKey := []string{"recentStreamInfo", username}
	if d := c.apiCache.Get(cacheKey); d != nil {
		return d.([2]string)[0], d.([2]string)[1], nil
	}

	id, err := c.GetIDForUsername(username)
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

	if err := c.request(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf("https://api.twitch.tv/helix/channels?broadcaster_id=%s", id),
		nil,
		&payload,
	); err != nil {
		return "", "", errors.Wrap(err, "request channel info")
	}

	if l := len(payload.Data); l != 1 {
		return "", "", errors.Errorf("unexpected number of users returned: %d", l)
	}

	// Stream-info can be changed at any moment, cache for a short period of time
	c.apiCache.Set(cacheKey, twitchMinCacheTime, [2]string{payload.Data[0].GameName, payload.Data[0].Title})

	return payload.Data[0].GameName, payload.Data[0].Title, nil
}

func (c Client) ModifyChannelInformation(ctx context.Context, broadcaster string, game, title *string) error {
	if game == nil && title == nil {
		return errors.New("netiher game nor title provided")
	}

	data := struct {
		GameID *string `json:"game_id,omitempty"`
		Title  *string `json:"title,omitempty"`
	}{
		Title: title,
	}

	if game != nil {
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
				if c.Name == *game {
					data.GameID = &c.ID
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
		c.request(ctx, http.MethodPost, fmt.Sprintf("https://api.twitch.tv/helix/channels?broadcaster_id=%s", broadcaster), body, nil),
		"executing request",
	)
}

func (c Client) request(ctx context.Context, method, url string, body io.Reader, out interface{}) error {
	log.WithFields(log.Fields{
		"method": method,
		"url":    url,
	}).Trace("Execute Twitch API request")

	return backoff.NewBackoff().WithMaxIterations(twitchRequestRetries).Retry(func() error {
		reqCtx, cancel := context.WithTimeout(ctx, twitchRequestTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(reqCtx, method, url, body)
		if err != nil {
			return errors.Wrap(err, "assemble request")
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Client-Id", c.clientID)
		req.Header.Set("Authorization", "Bearer "+c.token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrap(err, "execute request")
		}
		defer resp.Body.Close()

		if out == nil {
			return nil
		}

		return errors.Wrap(
			json.NewDecoder(resp.Body).Decode(out),
			"parse user info",
		)
	})
}
