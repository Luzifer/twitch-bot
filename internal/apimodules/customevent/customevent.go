package customevent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v2/pkg/database"
	"github.com/Luzifer/twitch-bot/v2/plugins"
)

const actorName = "customevent"

var (
	db               database.Connector
	eventCreatorFunc plugins.EventHandlerFunc
	formatMessage    plugins.MsgFormatter
	mc               *memoryCache

	ptrStringEmpty = func(s string) *string { return &s }("")
)

func Register(args plugins.RegistrationArguments) error {
	db = args.GetDatabaseConnector()
	if err := db.DB().AutoMigrate(&storedCustomEvent{}); err != nil {
		return errors.Wrap(err, "applying schema migration")
	}

	mc = &memoryCache{dbc: db}

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
			{
				Default:         "",
				Description:     "Time until the event is triggered (must be valid duration like 1h, 1h1m, 10s, ...)",
				Key:             "schedule_in",
				Name:            "Schedule In",
				Optional:        true,
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

	for schedule, fn := range map[string]func(){
		fmt.Sprintf("@every %s", cleanupTimeout): scheduleCleanup,
		"* * * * * *":                            scheduleSend,
	} {
		if _, err := args.RegisterCron(schedule, fn); err != nil {
			return errors.Wrap(err, "registering cron function")
		}
	}

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

func parseEvent(channel string, fieldData io.Reader) (*plugins.FieldCollection, error) {
	payload := make(map[string]any)

	if err := json.NewDecoder(fieldData).Decode(&payload); err != nil {
		return nil, errors.Wrap(err, "parsing event payload")
	}

	fields := plugins.FieldCollectionFromData(payload)
	fields.Set("channel", "#"+strings.TrimLeft(channel, "#"))

	return fields, nil
}

func triggerEvent(channel string, fieldData io.Reader) error {
	fields, err := parseEvent(channel, fieldData)
	if err != nil {
		return errors.Wrap(err, "parsing fields")
	}

	if err := eventCreatorFunc("custom", fields); err != nil {
		return errors.Wrap(err, "creating event")
	}

	return nil
}
