package customevent

import (
	"encoding/json"
	"net/http"

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
		Path:              "/create",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeNo200,
	})

	return nil
}

func handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	payload := make(map[string]any)

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, errors.Wrap(err, "parsing event payload").Error(), http.StatusBadRequest)
		return
	}

	if err := eventCreatorFunc("custom", plugins.FieldCollectionFromData(payload)); err != nil {
		http.Error(w, errors.Wrap(err, "creating event").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
