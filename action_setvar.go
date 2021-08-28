package main

import (
	"fmt"
	"net/http"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func() plugins.Actor { return &ActorSetVariable{} })

	registerRoute(plugins.HTTPRouteRegistrationArgs{
		Description:  "Returns the value as a plain string",
		HandlerFunc:  routeActorSetVarGetValue,
		Method:       http.MethodGet,
		Module:       "setvariable",
		Name:         "Get Variable Value",
		Path:         "/{name}",
		ResponseType: plugins.HTTPRouteResponseTypeTextPlain,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Name of the variable to query",
				Name:        "name",
			},
		},
	})

	registerRoute(plugins.HTTPRouteRegistrationArgs{
		Description: "Updates the value of the variable",
		HandlerFunc: routeActorSetVarSetValue,
		Method:      http.MethodPatch,
		Module:      "setvariable",
		Name:        "Set Variable Value",
		Path:        "/{name}",
		QueryParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Value to set for the given variable",
				Name:        "value",
				Required:    true,
				Type:        "string",
			},
		},
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Name of the variable to update",
				Name:        "name",
			},
		},
	})
}

type ActorSetVariable struct {
	Variable string `json:"variable" yaml:"variable"`
	Clear    bool   `json:"clear" yaml:"clear"`
	Set      string `json:"set" yaml:"set"`
}

func (a ActorSetVariable) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule) (preventCooldown bool, err error) {
	if a.Variable == "" {
		return false, nil
	}

	varName, err := formatMessage(a.Variable, m, r, nil)
	if err != nil {
		return false, errors.Wrap(err, "preparing variable name")
	}

	if a.Clear {
		return false, errors.Wrap(
			store.RemoveVariable(varName),
			"removing variable",
		)
	}

	value, err := formatMessage(a.Set, m, r, nil)
	if err != nil {
		return false, errors.Wrap(err, "preparing value")
	}

	return false, errors.Wrap(
		store.SetVariable(varName, value),
		"setting variable",
	)
}

func (a ActorSetVariable) IsAsync() bool { return false }
func (a ActorSetVariable) Name() string  { return "setvariable" }

func routeActorSetVarGetValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text-plain")
	fmt.Fprint(w, store.GetVariable(mux.Vars(r)["name"]))
}

func routeActorSetVarSetValue(w http.ResponseWriter, r *http.Request) {
	if err := store.SetVariable(mux.Vars(r)["name"], r.FormValue("value")); err != nil {
		http.Error(w, errors.Wrap(err, "updating value").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
