// Package messagehook contains actors to send discord / slack webhook
// requests
package messagehook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const (
	postTimeout = 5 * time.Second
)

var formatMessage plugins.MsgFormatter

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage

	discordActor{}.register(args)
	slackCompatibleActor{}.register(args)

	return nil
}

func sendPayload(hookURL string, payload any, expRespCode int) (preventCooldown bool, err error) {
	body := new(bytes.Buffer)
	if err = json.NewEncoder(body).Encode(payload); err != nil {
		return false, fmt.Errorf("marshalling payload: %w", err)
	}

	logrus.WithField("payload", body.String()).Trace("sending webhook payload")

	ctx, cancel := context.WithTimeout(context.Background(), postTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hookURL, body)
	if err != nil {
		return false, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing response body (leaked fd)")
		}
	}()

	if resp.StatusCode != expRespCode {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte(fmt.Errorf("reading body: %w", err).Error())
		}
		return false, fmt.Errorf("unexpected response code %d (Body: %s)", resp.StatusCode, body)
	}

	return false, nil
}
