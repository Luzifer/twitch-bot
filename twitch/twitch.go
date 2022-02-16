package twitch

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/go_helpers/v2/backoff"
)

const (
	timeDay = 24 * time.Hour

	twitchMinCacheTime = time.Second * 30

	twitchRequestRetries = 5
	twitchRequestTimeout = 2 * time.Second
)

const (
	authTypeUnauthorized authType = iota
	authTypeAppAccessToken
	authTypeBearerToken
)

type (
	Category struct {
		BoxArtURL string `json:"box_art_url"`
		ID        string `json:"id"`
		Name      string `json:"name"`
	}

	Client struct {
		clientID     string
		clientSecret string

		accessToken     string
		refreshToken    string
		tokenValidity   time.Time
		tokenUpdateHook func(string, string) error

		appAccessToken string

		apiCache *APICache
	}

	OAuthTokenResponse struct {
		AccessToken  string   `json:"access_token"`
		RefreshToken string   `json:"refresh_token"`
		ExpiresIn    int      `json:"expires_in"`
		Scope        []string `json:"scope"`
		TokenType    string   `json:"token_type"`
	}

	OAuthTokenValidationResponse struct {
		ClientID  string   `json:"client_id"`
		Login     string   `json:"login"`
		Scopes    []string `json:"scopes"`
		UserID    string   `json:"user_id"`
		ExpiresIn int      `json:"expires_in"`
	}

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
		TagIds       []string  `json:"tag_ids"`
		IsMature     bool      `json:"is_mature"`
	}

	User struct {
		DisplayName     string `json:"display_name"`
		ID              string `json:"id"`
		Login           string `json:"login"`
		ProfileImageURL string `json:"profile_image_url"`
	}

	authType uint8

	clientRequestOpts struct {
		AuthType        authType
		Body            io.Reader
		Context         context.Context
		Method          string
		NoRetry         bool
		NoValidateToken bool
		OKStatus        int
		Out             interface{}
		URL             string
	}
)

func New(clientID, clientSecret, accessToken, refreshToken string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,

		accessToken:  accessToken,
		refreshToken: refreshToken,

		apiCache: newTwitchAPICache(),
	}
}

func (c *Client) APICache() *APICache { return c.apiCache }

func (c *Client) GetAuthorizedUsername() (string, error) {
	var payload struct {
		Data []User `json:"data"`
	}

	if err := c.request(clientRequestOpts{
		AuthType: authTypeBearerToken,
		Context:  context.Background(),
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      "https://api.twitch.tv/helix/users",
	}); err != nil {
		return "", errors.Wrap(err, "request channel info")
	}

	if l := len(payload.Data); l != 1 {
		return "", errors.Errorf("unexpected number of users returned: %d", l)
	}

	return payload.Data[0].Login, nil
}

func (c *Client) GetDisplayNameForUser(username string) (string, error) {
	cacheKey := []string{"displayNameForUsername", username}
	if d := c.apiCache.Get(cacheKey); d != nil {
		return d.(string), nil
	}

	var payload struct {
		Data []User `json:"data"`
	}

	if err := c.request(clientRequestOpts{
		AuthType: authTypeAppAccessToken,
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

	if err := c.request(clientRequestOpts{
		AuthType: authTypeAppAccessToken,
		Context:  context.Background(),
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/users/follows?to_id=%s&from_id=%s", toID, fromID),
	}); err != nil {
		return time.Time{}, errors.Wrap(err, "request follow info")
	}

	if l := len(payload.Data); l != 1 {
		return time.Time{}, errors.Errorf("unexpected number of records returned: %d", l)
	}

	// Follow date will not change that often, cache for a long time
	c.apiCache.Set(cacheKey, timeDay, payload.Data[0].FollowedAt)

	return payload.Data[0].FollowedAt, nil
}

func (c *Client) GetToken() (string, error) {
	if err := c.ValidateToken(context.Background(), false); err != nil {
		if err = c.RefreshToken(); err != nil {
			return "", errors.Wrap(err, "refreshing token after validation error")
		}

		// Token was refreshed, therefore should now be valid
	}

	return c.accessToken, nil
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

	if err := c.request(clientRequestOpts{
		AuthType: authTypeAppAccessToken,
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

func (c *Client) SearchCategories(ctx context.Context, name string) ([]Category, error) {
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
		if err := c.request(clientRequestOpts{
			AuthType: authTypeBearerToken,
			Context:  ctx,
			Method:   http.MethodGet,
			OKStatus: http.StatusOK,
			Out:      &resp,
			URL:      fmt.Sprintf("https://api.twitch.tv/helix/search/categories?%s", params.Encode()),
		}); err != nil {
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

func (c *Client) HasLiveStream(username string) (bool, error) {
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

	if err := c.request(clientRequestOpts{
		AuthType: authTypeBearerToken,
		Context:  context.Background(),
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

func (c *Client) GetCurrentStreamInfo(username string) (*StreamInfo, error) {
	cacheKey := []string{"currentStreamInfo", username}
	if si := c.apiCache.Get(cacheKey); si != nil {
		return si.(*StreamInfo), nil
	}

	id, err := c.GetIDForUsername(username)
	if err != nil {
		return nil, errors.Wrap(err, "getting ID for username")
	}

	var payload struct {
		Data []*StreamInfo `json:"data"`
	}

	if err := c.request(clientRequestOpts{
		AuthType: authTypeBearerToken,
		Context:  context.Background(),
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/streams?user_id=%s", id),
	}); err != nil {
		return nil, errors.Wrap(err, "request channel info")
	}

	if l := len(payload.Data); l != 1 {
		return nil, errors.Errorf("unexpected number of users returned: %d", l)
	}

	// Stream-info can be changed at any moment, cache for a short period of time
	c.apiCache.Set(cacheKey, twitchMinCacheTime, payload.Data[0])

	return payload.Data[0], nil
}

func (c *Client) GetIDForUsername(username string) (string, error) {
	cacheKey := []string{"idForUsername", username}
	if d := c.apiCache.Get(cacheKey); d != nil {
		return d.(string), nil
	}

	var payload struct {
		Data []User `json:"data"`
	}

	if err := c.request(clientRequestOpts{
		AuthType: authTypeAppAccessToken,
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

func (c *Client) GetRecentStreamInfo(username string) (string, string, error) {
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

	if err := c.request(clientRequestOpts{
		AuthType: authTypeBearerToken,
		Context:  context.Background(),
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

func (c *Client) ModifyChannelInformation(ctx context.Context, broadcasterName string, game, title *string) error {
	if game == nil && title == nil {
		return errors.New("netiher game nor title provided")
	}

	broadcaster, err := c.GetIDForUsername(broadcasterName)
	if err != nil {
		return errors.Wrap(err, "getting ID for broadcaster name")
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
					gid := c.ID
					data.GameID = &gid
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
		c.request(clientRequestOpts{
			AuthType: authTypeBearerToken,
			Body:     body,
			Context:  ctx,
			Method:   http.MethodPatch,
			OKStatus: http.StatusOK,
			URL:      fmt.Sprintf("https://api.twitch.tv/helix/channels?broadcaster_id=%s", broadcaster),
		}),
		"executing request",
	)
}

func (c *Client) RefreshToken() error {
	if c.refreshToken == "" {
		return errors.New("no refresh token set")
	}

	params := make(url.Values)
	params.Set("client_id", c.clientID)
	params.Set("client_secret", c.clientSecret)
	params.Set("refresh_token", c.refreshToken)
	params.Set("grant_type", "refresh_token")

	var resp OAuthTokenResponse

	if err := c.request(clientRequestOpts{
		AuthType: authTypeUnauthorized,
		Context:  context.Background(),
		Method:   http.MethodPost,
		OKStatus: http.StatusOK,
		Out:      &resp,
		URL:      fmt.Sprintf("https://id.twitch.tv/oauth2/token?%s", params.Encode()),
	}); err != nil {
		// Retried refresh failed, wipe tokens
		c.UpdateToken("", "")
		if c.tokenUpdateHook != nil {
			if herr := c.tokenUpdateHook("", ""); herr != nil {
				log.WithError(err).Error("Unable to store token wipe after refresh failure")
			}
		}

		return errors.Wrap(err, "executing request")
	}

	c.UpdateToken(resp.AccessToken, resp.RefreshToken)
	c.tokenValidity = time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	log.WithField("expiry", c.tokenValidity).Trace("Access token refreshed")

	if c.tokenUpdateHook == nil {
		return nil
	}

	return errors.Wrap(c.tokenUpdateHook(resp.AccessToken, resp.RefreshToken), "calling token update hook")
}

func (c *Client) SetTokenUpdateHook(f func(string, string) error) {
	c.tokenUpdateHook = f
}

func (c *Client) UpdateToken(accessToken, refreshToken string) {
	c.accessToken = accessToken
	c.refreshToken = refreshToken
}

func (c *Client) ValidateToken(ctx context.Context, force bool) error {
	if c.tokenValidity.After(time.Now()) && !force {
		// We do have an expiration time and it's not expired
		// so we can assume we've checked the token and it should
		// still be valid.
		// NOTE(kahlers): In case of a token revokation this
		// assumption is invalid and will lead to failing requests

		return nil
	}

	if c.accessToken == "" {
		return errors.New("no access token present")
	}

	var resp OAuthTokenValidationResponse

	if err := c.request(clientRequestOpts{
		AuthType:        authTypeBearerToken,
		Context:         ctx,
		Method:          http.MethodGet,
		NoRetry:         true,
		NoValidateToken: true,
		OKStatus:        http.StatusOK,
		Out:             &resp,
		URL:             "https://id.twitch.tv/oauth2/validate",
	}); err != nil {
		return errors.Wrap(err, "executing request")
	}

	if resp.ClientID != c.clientID {
		return errors.New("token belongs to different app")
	}

	c.tokenValidity = time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	log.WithField("expiry", c.tokenValidity).Trace("Access token validated")

	return nil
}

func (c *Client) createEventSubSubscription(ctx context.Context, sub eventSubSubscription) (*eventSubSubscription, error) {
	var (
		buf  = new(bytes.Buffer)
		resp struct {
			Total      int64                  `json:"total"`
			Data       []eventSubSubscription `json:"data"`
			Pagination struct {
				Cursor string `json:"cursor"`
			} `json:"pagination"`
		}
	)

	if err := json.NewEncoder(buf).Encode(sub); err != nil {
		return nil, errors.Wrap(err, "assemble subscribe payload")
	}

	if err := c.request(clientRequestOpts{
		AuthType: authTypeAppAccessToken,
		Body:     buf,
		Context:  ctx,
		Method:   http.MethodPost,
		OKStatus: http.StatusAccepted,
		Out:      &resp,
		URL:      "https://api.twitch.tv/helix/eventsub/subscriptions",
	}); err != nil {
		return nil, errors.Wrap(err, "executing request")
	}

	return &resp.Data[0], nil
}

func (c *Client) deleteEventSubSubscription(ctx context.Context, id string) error {
	return errors.Wrap(c.request(clientRequestOpts{
		AuthType: authTypeAppAccessToken,
		Context:  ctx,
		Method:   http.MethodDelete,
		OKStatus: http.StatusNoContent,
		URL:      fmt.Sprintf("https://api.twitch.tv/helix/eventsub/subscriptions?id=%s", id),
	}), "executing request")
}

func (c *Client) getEventSubSubscriptions(ctx context.Context) ([]eventSubSubscription, error) {
	var (
		out    []eventSubSubscription
		params = make(url.Values)
		resp   struct {
			Total      int64                  `json:"total"`
			Data       []eventSubSubscription `json:"data"`
			Pagination struct {
				Cursor string `json:"cursor"`
			} `json:"pagination"`
		}
	)

	for {
		if err := c.request(clientRequestOpts{
			AuthType: authTypeAppAccessToken,
			Context:  ctx,
			Method:   http.MethodGet,
			OKStatus: http.StatusOK,
			Out:      &resp,
			URL:      fmt.Sprintf("https://api.twitch.tv/helix/eventsub/subscriptions?%s", params.Encode()),
		}); err != nil {
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

func (c *Client) getTwitchAppAccessToken() (string, error) {
	if c.appAccessToken != "" {
		return c.appAccessToken, nil
	}

	var rData struct {
		AccessToken  string        `json:"access_token"`
		RefreshToken string        `json:"refresh_token"`
		ExpiresIn    int           `json:"expires_in"`
		Scope        []interface{} `json:"scope"`
		TokenType    string        `json:"token_type"`
	}

	params := make(url.Values)
	params.Set("client_id", c.clientID)
	params.Set("client_secret", c.clientSecret)
	params.Set("grant_type", "client_credentials")

	u, _ := url.Parse("https://id.twitch.tv/oauth2/token")
	u.RawQuery = params.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), twitchRequestTimeout)
	defer cancel()

	if err := c.request(clientRequestOpts{
		AuthType: authTypeUnauthorized,
		Context:  ctx,
		Method:   http.MethodPost,
		OKStatus: http.StatusOK,
		Out:      &rData,
		URL:      u.String(),
	}); err != nil {
		return "", errors.Wrap(err, "fetching token response")
	}

	c.appAccessToken = rData.AccessToken
	return rData.AccessToken, nil
}

func (c *Client) request(opts clientRequestOpts) error {
	log.WithFields(log.Fields{
		"method": opts.Method,
		"url":    c.replaceSecrets(opts.URL),
	}).Trace("Execute Twitch API request")

	var retries uint64 = twitchRequestRetries
	if opts.Body != nil || opts.NoRetry {
		// Body must be read only once, do not retry
		retries = 1
	}

	return backoff.NewBackoff().WithMaxIterations(retries).Retry(func() error {
		reqCtx, cancel := context.WithTimeout(opts.Context, twitchRequestTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(reqCtx, opts.Method, opts.URL, opts.Body)
		if err != nil {
			return errors.Wrap(err, "assemble request")
		}
		req.Header.Set("Content-Type", "application/json")

		switch opts.AuthType {
		case authTypeUnauthorized:
			// Nothing to do

		case authTypeAppAccessToken:
			accessToken, err := c.getTwitchAppAccessToken()
			if err != nil {
				return errors.Wrap(err, "getting app-access-token")
			}

			req.Header.Set("Authorization", "Bearer "+accessToken)
			req.Header.Set("Client-Id", c.clientID)

		case authTypeBearerToken:
			accessToken := c.accessToken
			if !opts.NoValidateToken {
				accessToken, err = c.GetToken()
				if err != nil {
					return errors.Wrap(err, "getting bearer access token")
				}
			}

			req.Header.Set("Authorization", "Bearer "+accessToken)
			req.Header.Set("Client-Id", c.clientID)

		default:
			return errors.New("invalid auth type specified")
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrap(err, "execute request")
		}
		defer resp.Body.Close()

		if opts.AuthType == authTypeAppAccessToken && resp.StatusCode == http.StatusUnauthorized {
			// Seems our token was somehow revoked, clear the token and retry which will get a new token
			c.appAccessToken = ""
			return errors.New("app-access-token is invalid")
		}

		if opts.OKStatus != 0 && resp.StatusCode != opts.OKStatus {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return errors.Wrapf(err, "unexpected status %d and cannot read body", resp.StatusCode)
			}
			return errors.Errorf("unexpected status %d: %s", resp.StatusCode, body)
		}

		if opts.Out == nil {
			return nil
		}

		return errors.Wrap(
			json.NewDecoder(resp.Body).Decode(opts.Out),
			"parse user info",
		)
	})
}

func (c *Client) replaceSecrets(u string) string {
	var replacements []string

	for _, secret := range []string{
		c.accessToken,
		c.refreshToken,
		c.clientSecret,
	} {
		if secret == "" {
			continue
		}
		replacements = append(replacements, secret, c.hashSecret(secret))
	}

	return strings.NewReplacer(replacements...).Replace(u)
}

func (*Client) hashSecret(secret string) string {
	h := sha256.New()
	h.Write([]byte(secret))
	return fmt.Sprintf("[sha256:%x]", h.Sum(nil))
}
