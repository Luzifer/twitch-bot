package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func getCurrentTrackForChannel(channel string) (track string, err error) {
	channel = strings.TrimLeft(channel, "#")

	conf, err := oauthConfig(channel, "")
	if err != nil {
		return "", fmt.Errorf("getting oauth config: %w", err)
	}

	var token *oauth2.Token
	if err = db.ReadEncryptedCoreMeta(strings.Join([]string{"spotify-auth", channel}, ":"), &token); err != nil {
		return "", fmt.Errorf("loading oauth token: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), spotifyRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.spotify.com/v1/me/player/currently-playing", nil)
	if err != nil {
		return "", fmt.Errorf("creating currently-playing request: %w", err)
	}

	resp, err := conf.Client(context.Background(), token).Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing Spotify response body (leaked fd)")
		}
	}()

	defer func() {
		if err := db.StoreEncryptedCoreMeta(strings.Join([]string{"spotify-auth", channel}, ":"), token); err != nil {
			logrus.WithError(err).Error("storing back Spotify auth token")
		}
	}()

	var payload currentPlayingTrackResponse
	if err = json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	var artistNames []string
	for _, artist := range payload.Item.Artists {
		artistNames = append(artistNames, artist.Name)
	}

	return strings.Join([]string{
		strings.Join(artistNames, ", "),
		payload.Item.Name,
	}, " - "), nil
}
