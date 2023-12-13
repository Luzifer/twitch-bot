package twitch

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/go_helpers/v2/backoff"
)

const (
	timeDay = 24 * time.Hour

	tokenValidityRecheckInterval = time.Hour

	twitchMinCacheTime = time.Second * 30

	twitchRequestRetries = 5
	twitchRequestTimeout = 2 * time.Second
)

const (
	AuthTypeUnauthorized AuthType = iota
	AuthTypeAppAccessToken
	AuthTypeBearerToken
)

type (
	Client struct {
		clientID     string
		clientSecret string

		accessToken          string
		refreshToken         string
		tokenValidity        time.Time
		tokenValidityChecked time.Time
		tokenUpdateHook      func(string, string) error

		appAccessToken string

		apiCache *APICache
	}

	ErrorResponse struct {
		Error   string `json:"error"`
		Status  int    `json:"status"`
		Message string `json:"message"`
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

	AuthType uint8

	ClientRequestOpts struct {
		AuthType        AuthType
		Body            io.Reader
		Context         context.Context
		Method          string
		NoRetry         bool
		NoValidateToken bool
		OKStatus        int
		Out             interface{}
		URL             string
		ValidateFunc    func(ClientRequestOpts, *http.Response) error
	}
)

// ValidateStatus is the default validation function used when no
// ValidateFunc is given in the ClientRequestOpts and checks for the
// returned HTTP status is equal to the OKStatus.
//
// When the status is http.StatusTooManyRequests the function will
// return an error terminating any retries as retrying would not make
// sense (the error returned from Request will still be an HTTPError
// with status 429).
//
// When wrapping this function the body should not have been read
// before in order to have the response body available in the returned
// HTTPError
func ValidateStatus(opts ClientRequestOpts, resp *http.Response) error {
	if opts.OKStatus != 0 && resp.StatusCode != opts.OKStatus {
		// We shall not accept this!
		var ret error

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			ret = newHTTPError(resp.StatusCode, nil, err)
		} else {
			ret = newHTTPError(resp.StatusCode, body, nil)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			// Twitch doesn't want to hear any more of this
			return backoff.NewErrCannotRetry(ret)
		}
		return ret
	}

	return nil
}

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

func (c *Client) GetToken() (string, error) {
	if err := c.ValidateToken(context.Background(), false); err != nil {
		if err = c.RefreshToken(); err != nil {
			return "", errors.Wrap(err, "refreshing token after validation error")
		}

		// Token was refreshed, therefore should now be valid
	}

	return c.accessToken, nil
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

	err := c.Request(ClientRequestOpts{
		AuthType: AuthTypeUnauthorized,
		Context:  context.Background(),
		Method:   http.MethodPost,
		OKStatus: http.StatusOK,
		Out:      &resp,
		URL:      fmt.Sprintf("https://id.twitch.tv/oauth2/token?%s", params.Encode()),
	})
	switch {
	case err == nil:
		// That's fine, just continue

	case errors.Is(err, ErrAnyHTTPError):
		// Retried refresh failed, wipe tokens
		log.WithError(err).Warning("resetting tokens after refresh-failure")
		c.UpdateToken("", "")
		if c.tokenUpdateHook != nil {
			if herr := c.tokenUpdateHook("", ""); herr != nil {
				log.WithError(herr).Error("Unable to store token wipe after refresh failure")
			}
		}
		return errors.Wrap(err, "executing request")

	default:
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
	if c.tokenValidity.After(time.Now()) && time.Since(c.tokenValidityChecked) < tokenValidityRecheckInterval && !force {
		// We do have an expiration time and it's not expired
		// so we can assume we've checked the token and it should
		// still be valid.
		// To detect a token revokation early-ish we re-check the
		// token in defined interval. This is not the optimal
		// solution as we will get failing requests between revokation
		// and recheck but it's better than nothing.

		return nil
	}

	if c.accessToken == "" {
		return errors.New("no access token present")
	}

	var resp OAuthTokenValidationResponse

	if err := c.Request(ClientRequestOpts{
		AuthType:        AuthTypeBearerToken,
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
	c.tokenValidityChecked = time.Now()
	log.WithField("expiry", c.tokenValidity).Trace("Access token validated")

	return nil
}

func (c *Client) GetTwitchAppAccessToken() (string, error) {
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

	if err := c.Request(ClientRequestOpts{
		AuthType: AuthTypeUnauthorized,
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

//nolint:gocyclo // Not gonna split to keep as a logical unit
func (c *Client) Request(opts ClientRequestOpts) error {
	log.WithFields(log.Fields{
		"method": opts.Method,
		"url":    c.replaceSecrets(opts.URL),
	}).Trace("Execute Twitch API request")

	var retries uint64 = twitchRequestRetries
	if opts.Body != nil || opts.NoRetry {
		// Body must be read only once, do not retry
		retries = 1
	}

	if opts.ValidateFunc == nil {
		opts.ValidateFunc = ValidateStatus
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
		case AuthTypeUnauthorized:
			// Nothing to do

		case AuthTypeAppAccessToken:
			accessToken, err := c.GetTwitchAppAccessToken()
			if err != nil {
				return errors.Wrap(err, "getting app-access-token")
			}

			req.Header.Set("Authorization", "Bearer "+accessToken)
			req.Header.Set("Client-Id", c.clientID)

		case AuthTypeBearerToken:
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

		if opts.AuthType == AuthTypeAppAccessToken && resp.StatusCode == http.StatusUnauthorized {
			// Seems our token was somehow revoked, clear the token and retry which will get a new token
			c.appAccessToken = ""
			return errors.New("app-access-token is invalid")
		}

		if err = opts.ValidateFunc(opts, resp); err != nil {
			return err
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
