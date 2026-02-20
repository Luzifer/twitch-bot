// Package counter contains actors and template functions to work with
// database stored counters
package counter

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"gopkg.in/irc.v4"
	"gorm.io/gorm"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var (
	db            database.Connector
	formatMessage plugins.MsgFormatter

	errNotAValue = fmt.Errorf("not a value")
)

// Register provides the plugins.RegisterFunc
//
//nolint:funlen // This function is a few lines too long but only contains definitions
func Register(args plugins.RegistrationArguments) (err error) {
	db = args.GetDatabaseConnector()
	if err = db.DB().AutoMigrate(&counter{}); err != nil {
		return errors.Wrap(err, "applying schema migration")
	}

	args.RegisterCopyDatabaseFunc("counter", func(src, target *gorm.DB) error {
		return database.CopyObjects(src, target, &counter{})
	})

	formatMessage = args.FormatMessage

	args.RegisterActor("counter", func() plugins.Actor { return &actorCounter{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
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
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
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

	if err = args.RegisterAPIRoute(plugins.HTTPRouteRegistrationArgs{
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
	}); err != nil {
		return fmt.Errorf("registering API route: %w", err)
	}

	if err = args.RegisterAPIRoute(plugins.HTTPRouteRegistrationArgs{
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
		RequiresWriteAuth: true,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Name of the counter to update",
				Name:        "name",
			},
		},
	}); err != nil {
		return fmt.Errorf("registering API route: %w", err)
	}

	args.RegisterTemplateFunction("channelCounter", func(_ *irc.Message, _ *plugins.Rule, fields *fieldcollection.FieldCollection) interface{} {
		return func(name string) (string, error) {
			channel, err := fields.String("channel")
			if err != nil {
				return "", errors.Wrap(err, "channel not available")
			}

			return strings.Join([]string{channel, name}, ":"), nil
		}
	}, plugins.TemplateFuncDocumentation{
		Description: "Wraps the counter name into a channel specific counter name including the channel name",
		Syntax:      "channelCounter <counter name>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:       `{{ channelCounter "test" }}`,
			ExpectedOutput: "#example:test",
		},
	})

	args.RegisterTemplateFunction("counterRank", plugins.GenericTemplateFunctionGetter(func(prefix, name string) (res struct{ Rank, Count int64 }, err error) {
		res.Rank, res.Count, err = getCounterRank(db, prefix, name)
		return res, errors.Wrap(err, "getting counter rank")
	}), plugins.TemplateFuncDocumentation{
		Description: "Returns the rank of the given counter and the total number of counters in given counter prefix",
		Syntax:      `counterRank <prefix> <name>`,
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ $cr := counterRank (list .channel "test" "" | join ":") (list .channel "test" "foo" | join ":") }}{{ $cr.Rank }}/{{ $cr.Count }}`,
			FakedOutput: "2/6",
		},
	})

	args.RegisterTemplateFunction("counterTopList", plugins.GenericTemplateFunctionGetter(func(prefix string, n int, orderBy string) ([]counter, error) {
		return getCounterTopList(db, prefix, n, orderBy)
	}), plugins.TemplateFuncDocumentation{
		Description: "Returns the top n counters for the given prefix as objects with Name and Value fields. Can be ordered by `name` / `value` / `first_seen` / `last_modified` ascending (`ASC`) or descending (`DESC`): i.e. `last_modified DESC` defaults to `value DESC`",
		Syntax:      `counterTopList <prefix> <n> [orderBy]`,
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ range (counterTopList (list .channel "test" "" | join ":") 3) }}{{ .Name }}: {{ .Value }} - {{ end }}`,
			FakedOutput: "#example:test:foo: 5 - #example:test:bar: 4 - ",
		},
	})

	args.RegisterTemplateFunction("counterValue", plugins.GenericTemplateFunctionGetter(func(name string, _ ...string) (int64, error) {
		return getCounterValue(db, name)
	}), plugins.TemplateFuncDocumentation{
		Description: "Returns the current value of the counter which identifier was supplied",
		Syntax:      "counterValue <counter name>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ counterValue (list .channel "test" | join ":") }}`,
			FakedOutput: "5",
		},
	})

	args.RegisterTemplateFunction("counterValueAdd", plugins.GenericTemplateFunctionGetter(func(name string, val ...int64) (int64, error) {
		var mod int64 = 1
		if len(val) > 0 {
			mod = val[0]
		}

		if err := updateCounter(db, name, mod, false, time.Now()); err != nil {
			return 0, errors.Wrap(err, "updating counter")
		}

		return getCounterValue(db, name)
	}), plugins.TemplateFuncDocumentation{
		Description: "Adds the given value (or 1 if no value) to the counter and returns its new value",
		Syntax:      "counterValueAdd <counter name> [increase=1]",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ counterValueAdd "myCounter" }} {{ counterValueAdd "myCounter" 5 }}`,
			FakedOutput: "1 6",
		},
	})

	return nil
}

type actorCounter struct{}

func (a actorCounter) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	counterName, err := formatMessage(attrs.MustString("counter", nil), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "preparing response")
	}

	// First lets look whether we shall set the counter (counter_set is
	// defined and the template evaluates into something which is not
	// an empty string)
	counterSet, err := a.parseAttributeTemplateToNumber(m, r, eventData, attrs, "counter_set", 0)
	switch {
	case err == nil:
		// Nice, we got a value to set
		if err = updateCounter(db, counterName, counterSet, true, time.Now()); err != nil {
			return false, fmt.Errorf("setting counter: %w", err)
		}
		return false, nil

	case errors.Is(err, errNotAValue):
		// Nope, not a set but that's fine, we just go to step-adjustment

	default:
		// B0rked
		return false, fmt.Errorf("parsing counter-set: %w", err)
	}

	// Second check whether we do have a template in counter_step and it
	// evaluates into a non-empty string and then adjust the counter
	// accordingly
	counterStep, err := a.parseAttributeTemplateToNumber(m, r, eventData, attrs, "counter_step", 1)
	switch {
	case err == nil, errors.Is(err, errNotAValue):
		// Either got a value or there was none, therefore the default was
		// returned which is 1 and we can apply this
		if err = updateCounter(db, counterName, counterStep, false, time.Now()); err != nil {
			return false, fmt.Errorf("updating counter: %w", err)
		}
		return false, nil

	default:
		// B0rked
		return false, fmt.Errorf("parsing counter-step: %w", err)
	}
}

func (actorCounter) IsAsync() bool { return false }
func (actorCounter) Name() string  { return "counter" }

func (actorCounter) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "counter", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "counter_step", Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "counter_set", Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "counter", "counter_step", "counter_set"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

func (actorCounter) parseAttributeTemplateToNumber(
	m *irc.Message,
	r *plugins.Rule,
	eventData *fieldcollection.FieldCollection,
	attrs *fieldcollection.FieldCollection,
	field string,
	defaultValue int64,
) (v int64, err error) {
	// Get the string
	sv, err := attrs.String(field)
	switch {
	case err == nil:
		// We got a string and continue below

	case errors.Is(err, fieldcollection.ErrValueNotSet):
		// That's fine, the string is not available, we report that and
		// return the default value
		return defaultValue, errNotAValue

	default:
		// Not sure what brought us here but we should report that
		return defaultValue, fmt.Errorf("getting string value: %w", err)
	}

	// Now we need to evaluate the template
	sv, err = formatMessage(sv, m, r, eventData)
	if err != nil {
		return defaultValue, fmt.Errorf("executing template: %w", err)
	}

	// The template evaluated into an empty string, we don't try to
	// parse that and report it as a missing value with default
	if sv == "" {
		return defaultValue, errNotAValue
	}

	// The template was not empty, we need to parse the resulting int
	// and return it
	v, err = strconv.ParseInt(sv, 10, 64)
	if err != nil {
		return defaultValue, fmt.Errorf("parsing to int: %w", err)
	}

	return v, nil
}

func routeActorCounterGetValue(w http.ResponseWriter, r *http.Request) {
	template := r.FormValue("template")
	if template == "" {
		template = "%d"
	}

	cv, err := getCounterValue(db, mux.Vars(r)["name"])
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting value").Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text-plain")
	http.Error(w, fmt.Sprintf(template, cv), http.StatusOK)
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

	if err = updateCounter(db, mux.Vars(r)["name"], value, absolute, time.Now()); err != nil {
		http.Error(w, errors.Wrap(err, "updating value").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
