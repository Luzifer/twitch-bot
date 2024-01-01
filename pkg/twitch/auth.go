package twitch

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// GetTokenInfo requests a validation for the token set within the
// client and returns the authorized user, their granted scopes on this
// token and an error in case something went wrong.
func (c *Client) GetTokenInfo(ctx context.Context) (user string, scopes []string, expiresAt time.Time, err error) {
	var payload OAuthTokenValidationResponse

	if c.accessToken == "" {
		return "", nil, time.Time{}, errors.New("no access token present")
	}

	if err := c.Request(ctx, ClientRequestOpts{
		AuthType: AuthTypeBearerToken,
		Method:   http.MethodGet,
		OKStatus: http.StatusOK,
		Out:      &payload,
		URL:      "https://id.twitch.tv/oauth2/validate",
	}); err != nil {
		return "", nil, time.Time{}, errors.Wrap(err, "validating token")
	}

	return payload.Login, payload.Scopes, time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second), nil
}
