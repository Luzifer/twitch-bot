// Package messagehook contains actors to send discord / slack webhook
// requests
package messagehook

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
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
		return false, errors.Wrap(err, "marshalling payload")
	}

	logrus.WithField("payload", body.String()).Trace("sending webhook payload")

	ctx, cancel := context.WithTimeout(context.Background(), postTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hookURL, body)
	if err != nil {
		return false, errors.Wrap(err, "creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, errors.Wrap(err, "executing request")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing response body (leaked fd)")
		}
	}()

	if resp.StatusCode != expRespCode {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte(errors.Wrap(err, "reading body").Error())
		}
		return false, errors.Errorf("unexpected response code %d (Body: %s)", resp.StatusCode, body)
	}

	return false, nil
}
