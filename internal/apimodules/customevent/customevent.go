package customevent

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/plugins"
)

var eventCreatorFunc plugins.EventHandlerFunc

func Register(args plugins.RegistrationArguments) error {
	eventCreatorFunc = args.CreateEvent

	args.RegisterAPIRoute(plugins.HTTPRouteRegistrationArgs{
		Description:       "Creates an `custom` event containing the fields provided in the request body",
		HandlerFunc:       handleCreateEvent,
		Method:            http.MethodPost,
		Module:            "customevent",
		Name:              "Create custom event",
		Path:              "/{channel}",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeNo200,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Channel to create the event in",
				Name:        "channel",
			},
		},
	})

	return nil
}

func handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	var (
		channel = mux.Vars(r)["channel"]
		payload = make(map[string]any)
	)

	if channel == "" {
		http.Error(w, errors.New("missing channel").Error(), http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, errors.Wrap(err, "parsing event payload").Error(), http.StatusBadRequest)
		return
	}

	fields := plugins.FieldCollectionFromData(payload)
	fields.Set("channel", "#"+strings.TrimLeft(channel, "#"))

	if err := eventCreatorFunc("custom", fields); err != nil {
		http.Error(w, errors.Wrap(err, "creating event").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
