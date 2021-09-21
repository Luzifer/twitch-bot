package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func init() {
	registerAction("counter", func() plugins.Actor { return &ActorCounter{} })

	registerActorDocumentation(plugins.ActionDocumentation{
		Description: "Update counter values",
		Name:        "Modify Counter",
		Type:        "counter",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Name of the counter to update",
				Key:             "counter",
				Name:            "Counter",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "1",
				Description:     "Value to add to the counter",
				Key:             "counter_step",
				Name:            "Counter Step",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeInt64,
			},
			{
				Default:         "",
				Description:     "Value to set the counter to",
				Key:             "counter_set",
				Name:            "Counter Set",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	registerRoute(plugins.HTTPRouteRegistrationArgs{
		Description: "Returns the (formatted) value as a plain string",
		HandlerFunc: routeActorCounterGetValue,
		Method:      http.MethodGet,
		Module:      "counter",
		Name:        "Get Counter Value",
		Path:        "/{name}",
		QueryParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Template to apply to the value: Variations of %d sprintf template are supported once",
				Name:        "template",
				Required:    false,
				Type:        "string",
			},
		},
		ResponseType: plugins.HTTPRouteResponseTypeTextPlain,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Name of the counter to query",
				Name:        "name",
			},
		},
	})

	registerRoute(plugins.HTTPRouteRegistrationArgs{
		Description: "Updates the value of the counter",
		HandlerFunc: routeActorCounterSetValue,
		Method:      http.MethodPatch,
		Module:      "counter",
		Name:        "Set Counter Value",
		Path:        "/{name}",
		QueryParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "If set to `true` the given value is set instead of added",
				Name:        "absolute",
				Required:    false,
				Type:        "boolean",
			},
			{
				Description: "Value to add / set for the given counter",
				Name:        "value",
				Required:    true,
				Type:        "int64",
			},
		},
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Name of the counter to update",
				Name:        "name",
			},
		},
	})
}

type ActorCounter struct{}

func (a ActorCounter) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData plugins.FieldCollection, attrs plugins.FieldCollection) (preventCooldown bool, err error) {
	counterName, err := formatMessage(attrs.MustString("counter", nil), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "preparing response")
	}

	if counterSet := attrs.MustString("counter_set", ptrStringEmpty); counterSet != "" {
		parseValue, err := formatMessage(counterSet, m, r, eventData)
		if err != nil {
			return false, errors.Wrap(err, "execute counter value template")
		}

		counterValue, err := strconv.ParseInt(parseValue, 10, 64)
		if err != nil {
			return false, errors.Wrap(err, "parse counter value")
		}

		return false, errors.Wrap(
			store.UpdateCounter(counterName, counterValue, true),
			"set counter",
		)
	}

	var counterStep int64 = 1
	if s := attrs.MustInt64("counter_step", ptrIntZero); s != 0 {
		counterStep = s
	}

	return false, errors.Wrap(
		store.UpdateCounter(counterName, counterStep, false),
		"update counter",
	)
}

func (a ActorCounter) IsAsync() bool { return false }
func (a ActorCounter) Name() string  { return "counter" }

func (a ActorCounter) Validate(attrs plugins.FieldCollection) (err error) {
	if cn, err := attrs.String("counter"); err != nil || cn == "" {
		return errors.New("counter name must be non-empty string")
	}

	return nil
}

func routeActorCounterGetValue(w http.ResponseWriter, r *http.Request) {
	template := r.FormValue("template")
	if template == "" {
		template = "%d"
	}

	w.Header().Set("Content-Type", "text-plain")
	fmt.Fprintf(w, template, store.GetCounterValue(mux.Vars(r)["name"]))
}

func routeActorCounterSetValue(w http.ResponseWriter, r *http.Request) {
	var (
		absolute = r.FormValue("absolute") == "true"
		err      error
		value    int64
	)

	if value, err = strconv.ParseInt(r.FormValue("value"), 10, 64); err != nil {
		http.Error(w, errors.Wrap(err, "parsing value").Error(), http.StatusBadRequest)
		return
	}

	if err = store.UpdateCounter(mux.Vars(r)["name"], value, absolute); err != nil {
		http.Error(w, errors.Wrap(err, "updating value").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
