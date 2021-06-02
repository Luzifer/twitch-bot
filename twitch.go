package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Luzifer/go_helpers/v2/backoff"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	timeDay = 24 * time.Hour

	twitchRequestRetries = 5
	twitchRequestTimeout = 2 * time.Second
)

var twitch = newTwitchClient()

type twitchClient struct {
	apiCache *twitchAPICache
}

func newTwitchClient() *twitchClient {
	return &twitchClient{
		apiCache: newTwitchAPICache(),
	}
}

func (t twitchClient) getAuthorizedUsername() (string, error) {
	var payload struct {
		Data []struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"data"`
	}

	if err := t.request(
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

func (t twitchClient) GetDisplayNameForUser(username string) (string, error) {
	cacheKey := []string{"displayNameForUsername", username}
	if d := t.apiCache.Get(cacheKey); d != nil {
		return d.(string), nil
	}

	var payload struct {
		Data []struct {
			ID          string `json:"id"`
			DisplayName string `json:"display_name"`
			Login       string `json:"login"`
		} `json:"data"`
	}

	if err := t.request(
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
	t.apiCache.Set(cacheKey, time.Hour, payload.Data[0].DisplayName)

	return payload.Data[0].DisplayName, nil
}

func (t twitchClient) GetFollowDate(from, to string) (time.Time, error) {
	cacheKey := []string{"followDate", from, to}
	if d := t.apiCache.Get(cacheKey); d != nil {
		return d.(time.Time), nil
	}

	fromID, err := t.getIDForUsername(from)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "getting id for 'from' user")
	}
	toID, err := t.getIDForUsername(to)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "getting id for 'to' user")
	}

	var payload struct {
		Data []struct {
			FollowedAt time.Time `json:"followed_at"`
		} `json:"data"`
	}

	if err := t.request(
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
	t.apiCache.Set(cacheKey, timeDay, payload.Data[0].FollowedAt)

	return payload.Data[0].FollowedAt, nil
}

func (t twitchClient) HasLiveStream(username string) (bool, error) {
	cacheKey := []string{"hasLiveStream", username}
	if d := t.apiCache.Get(cacheKey); d != nil {
		return d.(bool), nil
	}

	var payload struct {
		Data []struct {
			ID        string `json:"id"`
			UserLogin string `json:"user_login"`
			Type      string `json:"type"`
		} `json:"data"`
	}

	if err := t.request(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf("https://api.twitch.tv/helix/streams?user_login=%s", username),
		nil,
		&payload,
	); err != nil {
		return false, errors.Wrap(err, "request stream info")
	}

	// Live status might change recently, cache for one minute
	t.apiCache.Set(cacheKey, time.Minute, len(payload.Data) == 1 && payload.Data[0].Type == "live")

	return len(payload.Data) == 1 && payload.Data[0].Type == "live", nil
}

func (t twitchClient) getIDForUsername(username string) (string, error) {
	cacheKey := []string{"idForUsername", username}
	if d := t.apiCache.Get(cacheKey); d != nil {
		return d.(string), nil
	}

	var payload struct {
		Data []struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"data"`
	}

	if err := t.request(
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
	t.apiCache.Set(cacheKey, timeDay, payload.Data[0].ID)

	return payload.Data[0].ID, nil
}

func (t twitchClient) GetRecentStreamInfo(username string) (string, string, error) {
	cacheKey := []string{"recentStreamInfo", username}
	if d := t.apiCache.Get(cacheKey); d != nil {
		return d.([2]string)[0], d.([2]string)[1], nil
	}

	id, err := t.getIDForUsername(username)
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

	if err := t.request(
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
	t.apiCache.Set(cacheKey, time.Minute, [2]string{payload.Data[0].GameName, payload.Data[0].Title})

	return payload.Data[0].GameName, payload.Data[0].Title, nil
}

func (twitchClient) request(ctx context.Context, method, url string, body io.Reader, out interface{}) error {
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
		req.Header.Set("Client-Id", cfg.TwitchClient)
		req.Header.Set("Authorization", "Bearer "+cfg.TwitchToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrap(err, "execute request")
		}
		defer resp.Body.Close()

		return errors.Wrap(
			json.NewDecoder(resp.Body).Decode(out),
			"parse user info",
		)
	})
}
