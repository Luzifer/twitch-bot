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
	registerAction(func() plugins.Actor { return &ActorCounter{} })

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

type ActorCounter struct {
	CounterSet  *string `json:"counter_set" yaml:"counter_set"`
	CounterStep *int64  `json:"counter_step" yaml:"counter_step"`
	Counter     *string `json:"counter" yaml:"counter"`
}

func (a ActorCounter) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData map[string]interface{}) (preventCooldown bool, err error) {
	if a.Counter == nil {
		return false, nil
	}

	counterName, err := formatMessage(*a.Counter, m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "preparing response")
	}

	if a.CounterSet != nil {
		parseValue, err := formatMessage(*a.CounterSet, m, r, eventData)
		if err != nil {
			return false, errors.Wrap(err, "execute counter value template")
		}

		counterValue, err := strconv.ParseInt(parseValue, 10, 64) //nolint:gomnd // Those numbers are static enough
		if err != nil {
			return false, errors.Wrap(err, "parse counter value")
		}

		return false, errors.Wrap(
			store.UpdateCounter(counterName, counterValue, true),
			"set counter",
		)
	}

	var counterStep int64 = 1
	if a.CounterStep != nil {
		counterStep = *a.CounterStep
	}

	return false, errors.Wrap(
		store.UpdateCounter(counterName, counterStep, false),
		"update counter",
	)
}

func (a ActorCounter) IsAsync() bool { return false }
func (a ActorCounter) Name() string  { return "counter" }

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

	if value, err = strconv.ParseInt(r.FormValue("value"), 10, 64); err != nil { //nolint:gomnd // Those numbers are static enough
		http.Error(w, errors.Wrap(err, "parsing value").Error(), http.StatusBadRequest)
		return
	}

	if err = store.UpdateCounter(mux.Vars(r)["name"], value, absolute); err != nil {
		http.Error(w, errors.Wrap(err, "updating value").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
