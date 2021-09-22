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
	registerAction("setvariable", func() plugins.Actor { return &ActorSetVariable{} })

	registerActorDocumentation(plugins.ActionDocumentation{
		Description: "Modify variable contents",
		Name:        "Modify Variable",
		Type:        "setvariable",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Name of the variable to update",
				Key:             "variable",
				Name:            "Variable",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "false",
				Description:     "Clear variable content and unset the variable",
				Key:             "clear",
				Name:            "Clear",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeBool,
			},
			{
				Default:         "",
				Description:     "Value to set the variable to",
				Key:             "set",
				Name:            "Set Content",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

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

type ActorSetVariable struct{}

func (a ActorSetVariable) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData plugins.FieldCollection, attrs plugins.FieldCollection) (preventCooldown bool, err error) {
	varName, err := formatMessage(attrs.MustString("variable", nil), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "preparing variable name")
	}

	if attrs.MustBool("clear", ptrBoolFalse) {
		return false, errors.Wrap(
			store.RemoveVariable(varName),
			"removing variable",
		)
	}

	value, err := formatMessage(attrs.MustString("set", ptrStringEmpty), m, r, eventData)
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

func (a ActorSetVariable) Validate(attrs plugins.FieldCollection) (err error) {
	if v, err := attrs.String("variable"); err != nil || v == "" {
		return errors.New("variable name must be non-empty string")
	}

	return nil
}

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
