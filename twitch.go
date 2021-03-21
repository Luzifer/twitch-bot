package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

var twitch = twitchClient{}

type twitchClient struct{}

func (t twitchClient) getAuthorizedUsername() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), twitchRequestTimeout)
	defer cancel()

	var payload struct {
		Data []struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"data"`
	}

	if err := t.request(ctx, http.MethodGet, "https://api.twitch.tv/helix/users", nil, &payload); err != nil {
		return "", errors.Wrap(err, "request channel info")
	}

	if l := len(payload.Data); l != 1 {
		return "", errors.Errorf("unexpected number of users returned: %d", l)
	}

	return payload.Data[0].Login, nil
}

func (t twitchClient) GetFollowDate(from, to string) (time.Time, error) {
	fromID, err := t.getIDForUsername(from)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "getting id for 'from' user")
	}
	toID, err := t.getIDForUsername(to)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "getting id for 'to' user")
	}

	ctx, cancel := context.WithTimeout(context.Background(), twitchRequestTimeout)
	defer cancel()

	var payload struct {
		Data []struct {
			FollowedAt time.Time `json:"followed_at"`
		} `json:"data"`
	}

	if err := t.request(
		ctx,
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

	return payload.Data[0].FollowedAt, nil
}

func (t twitchClient) getIDForUsername(username string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), twitchRequestTimeout)
	defer cancel()

	var payload struct {
		Data []struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"data"`
	}

	if err := t.request(
		ctx,
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

	return payload.Data[0].ID, nil
}

func (t twitchClient) GetRecentStreamInfo(username string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), twitchRequestTimeout)
	defer cancel()

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
		ctx,
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

	return payload.Data[0].GameName, payload.Data[0].Title, nil
}

func (twitchClient) request(ctx context.Context, method, url string, body io.Reader, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
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
}
