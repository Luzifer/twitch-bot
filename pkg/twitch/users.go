package twitch

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type (
	User struct {
		DisplayName     string `json:"display_name"`
		ID              string `json:"id"`
		Login           string `json:"login"`
		ProfileImageURL string `json:"profile_image_url"`
	}
)

var ErrUserDoesNotFollow = errors.New("no follow-relation found")

func (c *Client) GetAuthorizedUser() (userID string, userName string, err error) {
	var payload struct {
		Data []User `json:"data"`
	}

	if err := c.Request(ClientRequestOpts{
		AuthType: AuthTypeBearerToken,
		Context:  context.Background(),
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      "https://api.twitch.tv/helix/users",
	}); err != nil {
		return "", "", errors.Wrap(err, "request channel info")
	}

	if l := len(payload.Data); l != 1 {
		return "", "", errors.Errorf("unexpected number of users returned: %d", l)
	}

	return payload.Data[0].ID, payload.Data[0].Login, nil
}

func (c *Client) GetDisplayNameForUser(username string) (string, error) {
	cacheKey := []string{"displayNameForUsername", username}
	if d := c.apiCache.Get(cacheKey); d != nil {
		return d.(string), nil
	}

	var payload struct {
		Data []User `json:"data"`
	}

	if err := c.Request(ClientRequestOpts{
		AuthType: AuthTypeAppAccessToken,
		Context:  context.Background(),
		Method:   http.MethodGet,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/users?login=%s", username),
	}); err != nil {
		return "", errors.Wrap(err, "request channel info")
	}

	if l := len(payload.Data); l != 1 {
		return "", errors.Errorf("unexpected number of users returned: %d", l)
	}

	// The DisplayName for an username will not change (often), cache for a decent time
	c.apiCache.Set(cacheKey, time.Hour, payload.Data[0].DisplayName)

	return payload.Data[0].DisplayName, nil
}

func (c *Client) GetFollowDate(from, to string) (time.Time, error) {
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

	if err := c.Request(ClientRequestOpts{
		AuthType: AuthTypeAppAccessToken,
		Context:  context.Background(),
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/users/follows?to_id=%s&from_id=%s", toID, fromID),
	}); err != nil {
		return time.Time{}, errors.Wrap(err, "request follow info")
	}

	switch len(payload.Data) {
	case 0:
		return time.Time{}, ErrUserDoesNotFollow

	case 1:
		// Handled below, no error

	default:
		return time.Time{}, errors.Errorf("unexpected number of records returned: %d", len(payload.Data))
	}

	// Follow date will not change that often, cache for a long time
	c.apiCache.Set(cacheKey, timeDay, payload.Data[0].FollowedAt)

	return payload.Data[0].FollowedAt, nil
}

func (c *Client) GetIDForUsername(username string) (string, error) {
	cacheKey := []string{"idForUsername", username}
	if d := c.apiCache.Get(cacheKey); d != nil {
		return d.(string), nil
	}

	var payload struct {
		Data []User `json:"data"`
	}

	if err := c.Request(ClientRequestOpts{
		AuthType: AuthTypeAppAccessToken,
		Context:  context.Background(),
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/users?login=%s", username),
	}); err != nil {
		return "", errors.Wrap(err, "request channel info")
	}

	if l := len(payload.Data); l != 1 {
		return "", errors.Errorf("unexpected number of users returned: %d", l)
	}

	// The ID for an username will not change (often), cache for a long time
	c.apiCache.Set(cacheKey, timeDay, payload.Data[0].ID)

	return payload.Data[0].ID, nil
}

func (c *Client) GetUserInformation(user string) (*User, error) {
	var (
		out     User
		param   = "login"
		payload struct {
			Data []User `json:"data"`
		}
	)

	cacheKey := []string{"userInformation", user}
	if d := c.apiCache.Get(cacheKey); d != nil {
		out = d.(User)
		return &out, nil
	}

	if _, err := strconv.ParseInt(user, 10, 64); err == nil {
		param = "id"
	}

	if err := c.Request(ClientRequestOpts{
		AuthType: AuthTypeAppAccessToken,
		Context:  context.Background(),
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/users?%s=%s", param, user),
	}); err != nil {
		return nil, errors.Wrap(err, "request user info")
	}

	if l := len(payload.Data); l != 1 {
		return nil, errors.Errorf("unexpected number of records returned: %d", l)
	}

	// Follow date will not change that often, cache for a long time
	c.apiCache.Set(cacheKey, timeDay, payload.Data[0])
	out = payload.Data[0]

	return &out, nil
}
