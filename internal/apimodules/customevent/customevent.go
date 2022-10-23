package customevent

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v2/plugins"
)

const actorName = "customevent"

var (
	eventCreatorFunc plugins.EventHandlerFunc
	formatMessage    plugins.MsgFormatter
)

func Register(args plugins.RegistrationArguments) error {
	eventCreatorFunc = args.CreateEvent
	formatMessage = args.FormatMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Create a custom event",
		Name:        "Custom Event",
		Type:        actorName,

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "{}",
				Description:     "JSON representation of fields in the event (`map[string]any`)",
				Key:             "fields",
				Name:            "Fields",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

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
	channel := mux.Vars(r)["channel"]

	if channel == "" {
		http.Error(w, errors.New("missing channel").Error(), http.StatusBadRequest)
		return
	}
	channel = "#" + strings.TrimLeft(channel, "#") // Sanitize

	if err := triggerEvent(channel, r.Body); err != nil {
		http.Error(w, errors.Wrap(err, "creating event").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func triggerEvent(channel string, fieldData io.Reader) error {
	payload := make(map[string]any)

	if err := json.NewDecoder(fieldData).Decode(&payload); err != nil {
		return errors.Wrap(err, "parsing event payload")
	}

	fields := plugins.FieldCollectionFromData(payload)
	fields.Set("channel", "#"+strings.TrimLeft(channel, "#"))

	if err := eventCreatorFunc("custom", fields); err != nil {
		return errors.Wrap(err, "creating event")
	}

	return nil
}
