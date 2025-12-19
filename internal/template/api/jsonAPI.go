package api //revive:disable-line:var-naming

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/itchyny/gojq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	jqQueryTimeout       = 2 * time.Second
	remoteRequestTimeout = 5 * time.Second
)

func jsonAPI(uri, path string, fallback ...string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", errors.Wrap(err, "parsing URL")
	}

	query, err := gojq.Parse(path)
	if err != nil {
		return "", errors.Wrap(err, "parsing JSON path")
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

	execCtx, cancel := context.WithTimeout(context.Background(), jqQueryTimeout)
	defer cancel()

	data := make(map[string]any)
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", errors.Wrap(err, "parsing response JSON")
	}

	iter := query.RunWithContext(execCtx, data)
	v, ok := iter.Next()
	if !ok {
		if len(fallback) > 0 {
			return fallback[0], nil
		}

		return "", errors.New("no results found")
	}

	if err, ok := v.(error); ok {
		return "", errors.Wrap(err, "iterating path")
	}

	return fmt.Sprintf("%v", v), nil
}
