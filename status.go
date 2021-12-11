package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/plugins"
)

const statusIRCMessageReceivedTimeout = 5 * time.Minute

var statusIRCMessageReceived time.Time

type (
	statusResponse struct {
		Checks               []statusResponseCheck `json:"checks"`
		OverallStatusSuccess bool                  `json:"overall_status_success"`
	}

	statusResponseCheck struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Success     bool   `json:"success"`
		Error       string `json:"error,omitempty"`

		checkFn func() error
	}
)

func init() {
	registerRoute(plugins.HTTPRouteRegistrationArgs{
		HandlerFunc: handleStatusRequest,
		Method:      http.MethodGet,
		Module:      "status",
		Path:        "/status.json",
		QueryParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Set the response status for failing checks",
				Name:        "fail-status",
				Required:    false,
				Type:        "integer",
			},
		},
		RequiresWriteAuth: false,
		ResponseType:      plugins.HTTPRouteResponseTypeJSON,
	})
}

func handleStatusRequest(w http.ResponseWriter, r *http.Request) {
	failStatus := http.StatusInternalServerError
	if v, err := strconv.Atoi(r.FormValue("fail-status")); err == nil {
		failStatus = v
	}

	output := statusResponse{
		OverallStatusSuccess: true,
	}

	for _, chk := range []statusResponseCheck{
		{
			Name:        "IRC connection alive",
			Description: fmt.Sprintf("IRC connection received a message in last %s", statusIRCMessageReceivedTimeout),
			checkFn: func() error {
				if time.Since(statusIRCMessageReceived) > statusIRCMessageReceivedTimeout {
					return errors.New("message lifetime expired")
				}
				return nil
			},
		},
		{
			Name:        "Twitch Client Authorized",
			Description: "Twitch Client is authorized and can fetch authorized user",
			checkFn: func() error {
				if twitchClient == nil {
					return errors.New("not initialized")
				}

				_, err := twitchClient.GetAuthorizedUsername()
				return errors.Wrap(err, "fetching username")
			},
		},
	} {
		err := chk.checkFn()
		if err != nil {
			chk.Error = err.Error()
		}
		chk.Success = err == nil

		output.Checks = append(output.Checks, chk)
		output.OverallStatusSuccess = output.OverallStatusSuccess && chk.Success
	}

	w.Header().Set("Content-Type", "application/json")
	if !output.OverallStatusSuccess {
		w.WriteHeader(failStatus)
	}

	if err := json.NewEncoder(w).Encode(output); err != nil {
		http.Error(w, errors.Wrap(err, "encoding output").Error(), http.StatusInternalServerError)
		return
	}
}
