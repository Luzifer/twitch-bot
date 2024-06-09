// Package msgformat contains an API route to utilize the internal
// message formatter to format strings
package msgformat

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var formatMessage plugins.MsgFormatter

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) (err error) {
	formatMessage = args.FormatMessage

	if err = args.RegisterAPIRoute(plugins.HTTPRouteRegistrationArgs{
		Description: "Takes the given template and renders it using the same renderer as messages in the channel are",
		HandlerFunc: handleFormattedMessage,
		Method:      http.MethodGet,
		Module:      "msgformat",
		Name:        "Format Message",
		Path:        "/format",
		QueryParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "The template to execute",
				Name:        "template",
				Required:    true,
				Type:        "string",
			},
		},
		RequiresWriteAuth: true, // This module can potentially be used to harvest data / read internal variables so it is handled as a write-module
		ResponseType:      plugins.HTTPRouteResponseTypeTextPlain,
	}); err != nil {
		return fmt.Errorf("registering API route: %w", err)
	}

	return nil
}

func handleFormattedMessage(w http.ResponseWriter, r *http.Request) {
	tpl := r.FormValue("template")
	if tpl == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	msg, err := formatMessage(tpl, nil, nil, nil)
	if err != nil {
		http.Error(w, errors.Wrap(err, "executing template").Error(), http.StatusInternalServerError)
		return
	}

	http.Error(w, msg, http.StatusOK)
}
