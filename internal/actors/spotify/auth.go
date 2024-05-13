package spotify

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Luzifer/twitch-bot/v3/internal/locker"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const expiryGrace = 10 * time.Second

func getAuthorizedClient(channel, redirectURL string) (client *http.Client, err error) {
	// In templating functions are called multiple times at once which
	// with Spotify replacing the refresh-token on each renew would kill
	// the stored token when multiple spotify functions are called at
	// once. Therefore we do have this method locking itself until it
	// has successfully made one request to the users profile and therefore
	// renewed the token. The next request then will use the token the
	// previous request renewed.
	locker.LockByKey(strings.Join([]string{"spotify", "api-access", channel}, ":"))
	defer locker.UnlockByKey(strings.Join([]string{"spotify", "api-access", channel}, ":"))

	conf, err := oauthConfig(channel, redirectURL)
	if err != nil {
		return nil, fmt.Errorf("getting oauth config: %w", err)
	}

	var token *oauth2.Token
	if err = db.ReadEncryptedCoreMeta(strings.Join([]string{"spotify-auth", channel}, ":"), &token); err != nil {
		return nil, fmt.Errorf("loading oauth token: %w", err)
	}

	ts := conf.TokenSource(context.Background(), token)

	if token.Expiry.After(time.Now().Add(expiryGrace)) {
		// Token is still valid long enough, we spare the resources to do
		// the profile fetch and directly return the client with the token
		// as the scenario described here does not apply.
		return oauth2.NewClient(context.Background(), ts), nil
	}

	logrus.WithField("channel", channel).Debug("refreshing spotify token")

	ctx, cancel := context.WithTimeout(context.Background(), spotifyRequestTimeout)
	defer cancel()

	// We do a request to /me once to refresh the token if needed
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.spotify.com/v1/me", nil)
	if err != nil {
		return nil, fmt.Errorf("creating currently-playing request: %w", err)
	}

	oauthClient := oauth2.NewClient(context.Background(), ts)

	resp, err := oauthClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing Spotify response body (leaked fd)")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("requesting user profile: %w", err)
	}

	updToken, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("getting updated token: %w", err)
	}

	if err := db.StoreEncryptedCoreMeta(strings.Join([]string{"spotify-auth", channel}, ":"), updToken); err != nil {
		logrus.WithError(err).Error("storing back Spotify auth token")
	}

	return oauthClient, nil
}

func oauthConfig(channel, redirectURL string) (conf *oauth2.Config, err error) {
	clientID, err := getModuleConfig(actorName, channel).String("clientId")
	if err != nil {
		return nil, fmt.Errorf("getting clientId for channel: %w", err)
	}

	return &oauth2.Config{
		ClientID: clientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
		RedirectURL: redirectURL,
		Scopes:      []string{"user-read-currently-playing"},
	}, nil
}
