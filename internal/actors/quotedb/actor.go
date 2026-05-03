// Package quotedb contains a quote database and actor / api methods
// to manage it
package quotedb

import (
	"fmt"
	"strconv"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"gopkg.in/irc.v4"
	"gorm.io/gorm"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const (
	actorName = "quotedb"
)

type (
	actor struct{}
)

var (
	db            database.Connector
	formatMessage plugins.MsgFormatter
	send          plugins.SendMessageFunc

	// ptrStringEmpty     = func(v string) *string { return &v }("")
	// ptrStringOutFormat = func(v string) *string { return &v }("Quote #{{ .index }}: {{ .quote }}")
	// ptrStringZero      = func(v string) *string { return &v }("0")
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) (err error) {
	db = args.GetDatabaseConnector()
	if err = db.DB().AutoMigrate(&quote{}); err != nil {
		return fmt.Errorf("applying schema migration: %w", err)
	}

	args.RegisterCopyDatabaseFunc("quote", func(src, target *gorm.DB) error {
		return database.CopyObjects(src, target, &quote{})
	})

	formatMessage = args.FormatMessage
	send = args.SendMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Manage a database of quotes in your channel",
		Name:        "Quote Database",
		Type:        actorName,

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Action to execute (one of: add, del, get)",
				Key:             "action",
				Name:            "Action",
				Optional:        false,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "0",
				Description:     "Index of the quote to work with, must yield a number (required on 'del', optional on 'get')",
				Key:             "index",
				Name:            "Index",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "Quote to add: Format like you like your quote, nothing is added (required on: add)",
				Key:             "quote",
				Name:            "Quote",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "Quote #{{ .index }}: {{ .quote }}",
				Description:     "Format to use when posting a quote (required on: get)",
				Key:             "format",
				Name:            "Format",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	if err = registerAPI(args.RegisterAPIRoute); err != nil {
		return fmt.Errorf("registering API: %w", err)
	}

	args.RegisterTemplateFunction("lastQuoteIndex", func(m *irc.Message, _ *plugins.Rule, _ *fieldcollection.FieldCollection) any {
		return func() (int, error) {
			return getMaxQuoteIdx(db, plugins.DeriveChannel(m, nil))
		}
	}, plugins.TemplateFuncDocumentation{
		Description: "Gets the last quote index in the quote database for the current channel",
		Syntax:      "lastQuoteIndex",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `Last Quote: #{{ lastQuoteIndex }}`,
			FakedOutput: "Last Quote: #32",
		},
	})

	return nil
}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	var (
		action   = attrs.MustString("action", helpers.Ptr(""))
		indexStr = attrs.MustString("index", helpers.Ptr("0"))
		quote    = attrs.MustString("quote", helpers.Ptr(""))
	)

	if indexStr == "" {
		indexStr = "0"
	}

	if indexStr, err = formatMessage(indexStr, m, r, eventData); err != nil {
		return false, fmt.Errorf("formatting index: %w", err)
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return false, fmt.Errorf("parsing index to number: %w", err)
	}

	switch action {
	case "add":
		quote, err = formatMessage(quote, m, r, eventData)
		if err != nil {
			return false, fmt.Errorf("formatting quote: %w", err)
		}

		if err = addQuote(db, plugins.DeriveChannel(m, eventData), quote); err != nil {
			return false, fmt.Errorf("adding quote: %w", err)
		}

		return false, nil

	case "del":
		if err = delQuote(db, plugins.DeriveChannel(m, eventData), index); err != nil {
			return false, fmt.Errorf("storing quote database: %w", err)
		}

		return false, nil

	case "get":
		idx, quote, err := getQuote(db, plugins.DeriveChannel(m, eventData), index)
		if err != nil {
			return false, fmt.Errorf("getting quote: %w", err)
		}

		if idx == 0 {
			// No quote was found for the given idx
			return false, nil
		}

		fields := eventData.Clone()
		fields.Set("index", idx)
		fields.Set("quote", quote)

		format := attrs.MustString("format", helpers.Ptr("Quote #{{ .index }}: {{ .quote }}"))
		msg, err := formatMessage(format, m, r, fields)
		if err != nil {
			return false, fmt.Errorf("formatting output message: %w", err)
		}

		if err = send(&irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				plugins.DeriveChannel(m, eventData),
				msg,
			},
		}); err != nil {
			return false, fmt.Errorf("sending command: %w", err)
		}

		return false, nil
	}

	return false, nil
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "action", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "quote", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "index", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "format", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "index", "quote", "format"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	action := attrs.MustString("action", helpers.Ptr(""))

	switch action {
	case "add":
		if v, err := attrs.String("quote"); err != nil || v == "" {
			return fmt.Errorf("quote must be non-empty string for action add")
		}

	case "del":
		if v, err := attrs.String("index"); err != nil || v == "" {
			return fmt.Errorf("index must be non-empty string for adction del")
		}

	case "get":
		// No requirements

	default:
		return fmt.Errorf("action must be one of add, del or get")
	}

	return nil
}
