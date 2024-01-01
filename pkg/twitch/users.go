package twitch

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type (
	// User represents the data known about an user
	User struct {
		DisplayName     string `json:"display_name"`
		ID              string `json:"id"`
		Login           string `json:"login"`
		ProfileImageURL string `json:"profile_image_url"`
	}
)

// ErrUserDoesNotFollow states the user does not follow the given channel
var ErrUserDoesNotFollow = errors.New("no follow-relation found")

// GetAuthorizedUser returns the userID / userName of the user the
// client is authorized for
func (c *Client) GetAuthorizedUser(ctx context.Context) (userID string, userName string, err error) {
	var payload struct {
		Data []User `json:"data"`
	}

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeBearerToken,
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

// GetDisplayNameForUser returns the display name for a login name set
// by the user themselves
func (c *Client) GetDisplayNameForUser(ctx context.Context, username string) (string, error) {
	username = strings.TrimLeft(username, "#@")

	cacheKey := []string{"displayNameForUsername", username}
	if d := c.apiCache.Get(cacheKey); d != nil {
		return d.(string), nil
	}

	var payload struct {
		Data []User `json:"data"`
	}

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeAppAccessToken,
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

// GetFollowDate returns the point-in-time the {from} followed the {to}
// or an ErrUserDoesNotFollow in case they do not follow
func (c *Client) GetFollowDate(ctx context.Context, from, to string) (time.Time, error) {
	cacheKey := []string{"followDate", from, to}
	if d := c.apiCache.Get(cacheKey); d != nil {
		return d.(time.Time), nil
	}

	fromID, err := c.GetIDForUsername(ctx, from)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "getting id for 'from' user")
	}
	toID, err := c.GetIDForUsername(ctx, to)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "getting id for 'to' user")
	}

	var payload struct {
		Data []struct {
			FollowedAt time.Time `json:"followed_at"`
		} `json:"data"`
	}

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeBearerToken,
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/channels/followers?broadcaster_id=%s&user_id=%s", toID, fromID),
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

// GetIDForUsername takes a login name and returns the userID for that
// username
func (c *Client) GetIDForUsername(ctx context.Context, username string) (string, error) {
	username = strings.TrimLeft(username, "#@")

	cacheKey := []string{"idForUsername", username}
	if d := c.apiCache.Get(cacheKey); d != nil {
		return d.(string), nil
	}

	var payload struct {
		Data []User `json:"data"`
	}

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeAppAccessToken,
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

// GetUsernameForID retrieves the login name (not the display name)
// for the given user ID
func (c *Client) GetUsernameForID(ctx context.Context, id string) (string, error) {
	cacheKey := []string{"usernameForID", id}
	if d := c.apiCache.Get(cacheKey); d != nil {
		return d.(string), nil
	}

	var payload struct {
		Data []User `json:"data"`
	}

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeAppAccessToken,
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/users?id=%s", id),
	}); err != nil {
		return "", errors.Wrap(err, "request channel info")
	}

	if l := len(payload.Data); l != 1 {
		return "", errors.Errorf("unexpected number of users returned: %d", l)
	}

	// The username for an ID will not change (often), cache for a long time
	c.apiCache.Set(cacheKey, timeDay, payload.Data[0].Login)

	return payload.Data[0].Login, nil
}

// GetUserInformation takes an userID or an userName and returns the
// User information for them
func (c *Client) GetUserInformation(ctx context.Context, user string) (*User, error) {
	user = strings.TrimLeft(user, "#@")

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

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeAppAccessToken,
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

	// User info will not change that often, cache for a long time
	c.apiCache.Set(cacheKey, timeDay, payload.Data[0])
	out = payload.Data[0]

	return &out, nil
}
