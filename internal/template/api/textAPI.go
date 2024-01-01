package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func textAPI(uri string, fallback ...string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", errors.Wrap(err, "parsing URL")
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), remoteRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", errors.Wrap(err, "assembling request")
	}
	req.Header.Set("User-Agent", "Luzifer/twitch-bot template/api/jsonAPI (https://github.com/Luzifer/twitch-bot)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "executing request")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing response body (leaked fd)")
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		// That's what we wanna see

	case http.StatusNoContent:
		if len(fallback) > 0 {
			return fallback[0], nil
		}
		return "", errors.Errorf("unexpected HTTP status %d without fallback", resp.StatusCode)

	default:
		return "", errors.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "reading response body")
	}

	if len(bytes.TrimSpace(content)) == 0 && len(fallback) > 0 {
		return fallback[0], nil
	}

	return string(content), nil
}
