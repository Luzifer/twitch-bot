package quotedb

import (
	"strconv"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/pkg/database"
	"github.com/Luzifer/twitch-bot/plugins"
)

const (
	actorName  = "quotedb"
	moduleUUID = "917c83ee-ed40-41e4-a558-1c2e59fdf1f5"
)

var (
	db            database.Connector
	formatMessage plugins.MsgFormatter

	ptrStringEmpty     = func(v string) *string { return &v }("")
	ptrStringOutFormat = func(v string) *string { return &v }("Quote #{{ .index }}: {{ .quote }}")
	ptrStringZero      = func(v string) *string { return &v }("0")
)

func Register(args plugins.RegistrationArguments) error {
	db = args.GetDatabaseConnector()
	if err := db.Migrate(actorName, database.NewEmbedFSMigrator(schema, "schema")); err != nil {
		return errors.Wrap(err, "applying schema migration")
	}

	formatMessage = args.FormatMessage

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

	registerAPI(args.RegisterAPIRoute)

	args.RegisterTemplateFunction("lastQuoteIndex", func(m *irc.Message, r *plugins.Rule, fields *plugins.FieldCollection) interface{} {
		return func() (int, error) {
			return getMaxQuoteIdx(plugins.DeriveChannel(m, nil))
		}
	})

	return nil
}

type (
	actor struct{}
)

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	var (
		action   = attrs.MustString("action", ptrStringEmpty)
		indexStr = attrs.MustString("index", ptrStringZero)
		quote    = attrs.MustString("quote", ptrStringEmpty)
	)

	if indexStr == "" {
		indexStr = "0"
	}

	if indexStr, err = formatMessage(indexStr, m, r, eventData); err != nil {
		return false, errors.Wrap(err, "formatting index")
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return false, errors.Wrap(err, "parsing index to number")
	}

	switch action {
	case "add":
		quote, err = formatMessage(quote, m, r, eventData)
		if err != nil {
			return false, errors.Wrap(err, "formatting quote")
		}

		return false, errors.Wrap(
			addQuote(plugins.DeriveChannel(m, eventData), quote),
			"adding quote",
		)

	case "del":
		return false, errors.Wrap(
			delQuote(plugins.DeriveChannel(m, eventData), index),
			"storing quote database",
		)

	case "get":
		idx, quote, err := getQuote(plugins.DeriveChannel(m, eventData), index)
		if err != nil {
			return false, errors.Wrap(err, "getting quote")
		}

		if idx == 0 {
			// No quote was found for the given idx
			return false, nil
		}

		fields := eventData.Clone()
		fields.Set("index", idx)
		fields.Set("quote", quote)

		format := attrs.MustString("format", ptrStringOutFormat)
		msg, err := formatMessage(format, m, r, fields)
		if err != nil {
			return false, errors.Wrap(err, "formatting output message")
		}

		return false, errors.Wrap(
			c.WriteMessage(&irc.Message{
				Command: "PRIVMSG",
				Params: []string{
					plugins.DeriveChannel(m, eventData),
					msg,
				},
			}),
			"sending command",
		)
	}

	return false, nil
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(attrs *plugins.FieldCollection) (err error) {
	action := attrs.MustString("action", ptrStringEmpty)

	switch action {
	case "add":
		if v, err := attrs.String("quote"); err != nil || v == "" {
			return errors.New("quote must be non-empty string for action add")
		}

	case "del":
		if v, err := attrs.String("index"); err != nil || v == "" {
			return errors.New("index must be non-empty string for adction del")
		}

	case "get":
		// No requirements

	default:
		return errors.New("action must be one of add, del or get")
	}

	return nil
}
