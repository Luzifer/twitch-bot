package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

var errNotPlaying = errors.New("nothing playing")

func getCurrentTrackForChannel(channel string) (track currentPlayingTrackResponse, err error) {
	channel = strings.TrimLeft(channel, "#")

	client, err := getAuthorizedClient(channel, "")
	if err != nil {
		return track, fmt.Errorf("retrieving authorized Spotify client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), spotifyRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.spotify.com/v1/me/player/currently-playing", nil)
	if err != nil {
		return track, fmt.Errorf("creating currently-playing request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return track, fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing Spotify response body (leaked fd)")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return track, fmt.Errorf("reading response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// This is perfect, continue below

	case http.StatusNoContent:
		// User is not playing anything
		return track, errNotPlaying

	case http.StatusUnauthorized:
		// The token is FUBAR
		return track, fmt.Errorf("token expired (HTTP 401 - unauthorized)")

	case http.StatusForbidden:
		// The request is FUBAR
		return track, fmt.Errorf("bad oAuth request, report this to dev (HTTP 403 - forbidden): %q", body)

	case http.StatusTooManyRequests:
		// We asked too often
		return track, fmt.Errorf("rate-limited (HTTP 429 - too many requests)")

	default:
		// WTF?
		return track, fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}

	if err = json.Unmarshal(body, &track); err != nil {
		return track, fmt.Errorf("decoding response (%q): %w", body, err)
	}

	return track, nil
}

func getCurrentArtistTitleForChannel(channel string) (artistTitle string, err error) {
	track, err := getCurrentTrackForChannel(channel)
	if err != nil {
		if errors.Is(err, errNotPlaying) {
			return "", nil
		}

		return "", fmt.Errorf("getting track info: %w", err)
	}

	var artistNames []string
	for _, artist := range track.Item.Artists {
		artistNames = append(artistNames, artist.Name)
	}

	return strings.Join([]string{
		strings.Join(artistNames, ", "),
		track.Item.Name,
	}, " - "), nil
}

func getCurrentLinkForChannel(channel string) (link string, err error) {
	track, err := getCurrentTrackForChannel(channel)
	if err != nil {
		if errors.Is(err, errNotPlaying) {
			return "", nil
		}

		return "", fmt.Errorf("getting track info: %w", err)
	}

	return track.Item.ExternalUrls.Spotify, nil
}
