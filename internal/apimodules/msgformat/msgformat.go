package msgformat

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var formatMessage plugins.MsgFormatter

func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage

	args.RegisterAPIRoute(plugins.HTTPRouteRegistrationArgs{
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
	})

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

	fmt.Fprint(w, msg)
}
