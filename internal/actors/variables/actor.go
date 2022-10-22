package variables

import (
	"fmt"
	"net/http"

	"github.com/go-irc/irc"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/pkg/database"
	"github.com/Luzifer/twitch-bot/plugins"
)

var (
	db            database.Connector
	formatMessage plugins.MsgFormatter

	ptrBoolFalse   = func(b bool) *bool { return &b }(false)
	ptrStringEmpty = func(s string) *string { return &s }("")
)

func Register(args plugins.RegistrationArguments) error {
	db = args.GetDatabaseConnector()
	if err := db.DB().AutoMigrate(&variable{}); err != nil {
		return errors.Wrap(err, "applying schema migration")
	}

	formatMessage = args.FormatMessage

	args.RegisterActor("setvariable", func() plugins.Actor { return &ActorSetVariable{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
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

	args.RegisterAPIRoute(plugins.HTTPRouteRegistrationArgs{
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

	args.RegisterAPIRoute(plugins.HTTPRouteRegistrationArgs{
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
		RequiresWriteAuth: true,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Name of the variable to update",
				Name:        "name",
			},
		},
	})

	args.RegisterTemplateFunction("variable", plugins.GenericTemplateFunctionGetter(func(name string, defVal ...string) (string, error) {
		value, err := GetVariable(db, name)
		if err != nil {
			return "", errors.Wrap(err, "getting variable")
		}

		if value == "" && len(defVal) > 0 {
			return defVal[0], nil
		}
		return value, nil
	}))

	return nil
}

type ActorSetVariable struct{}

func (a ActorSetVariable) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	varName, err := formatMessage(attrs.MustString("variable", nil), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "preparing variable name")
	}

	if attrs.MustBool("clear", ptrBoolFalse) {
		return false, errors.Wrap(
			RemoveVariable(db, varName),
			"removing variable",
		)
	}

	value, err := formatMessage(attrs.MustString("set", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "preparing value")
	}

	return false, errors.Wrap(
		SetVariable(db, varName, value),
		"setting variable",
	)
}

func (a ActorSetVariable) IsAsync() bool { return false }
func (a ActorSetVariable) Name() string  { return "setvariable" }

func (a ActorSetVariable) Validate(attrs *plugins.FieldCollection) (err error) {
	if v, err := attrs.String("variable"); err != nil || v == "" {
		return errors.New("variable name must be non-empty string")
	}

	return nil
}

func routeActorSetVarGetValue(w http.ResponseWriter, r *http.Request) {
	vc, err := GetVariable(db, mux.Vars(r)["name"])
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting value").Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text-plain")
	fmt.Fprint(w, vc)
}

func routeActorSetVarSetValue(w http.ResponseWriter, r *http.Request) {
	if err := SetVariable(db, mux.Vars(r)["name"], r.FormValue("value")); err != nil {
		http.Error(w, errors.Wrap(err, "updating value").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
