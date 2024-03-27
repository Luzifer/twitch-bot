// Package quotedb contains a quote database and actor / api methods
// to manage it
package quotedb

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"
	"gorm.io/gorm"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const (
	actorName = "quotedb"
)

var (
	db            database.Connector
	formatMessage plugins.MsgFormatter
	send          plugins.SendMessageFunc

	ptrStringEmpty     = func(v string) *string { return &v }("")
	ptrStringOutFormat = func(v string) *string { return &v }("Quote #{{ .index }}: {{ .quote }}")
	ptrStringZero      = func(v string) *string { return &v }("0")
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) (err error) {
	db = args.GetDatabaseConnector()
	if err = db.DB().AutoMigrate(&quote{}); err != nil {
		return errors.Wrap(err, "applying schema migration")
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

	args.RegisterTemplateFunction("lastQuoteIndex", func(m *irc.Message, _ *plugins.Rule, _ *plugins.FieldCollection) interface{} {
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

type (
	actor struct{}
)

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
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
			addQuote(db, plugins.DeriveChannel(m, eventData), quote),
			"adding quote",
		)

	case "del":
		return false, errors.Wrap(
			delQuote(db, plugins.DeriveChannel(m, eventData), index),
			"storing quote database",
		)

	case "get":
		idx, quote, err := getQuote(db, plugins.DeriveChannel(m, eventData), index)
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
			send(&irc.Message{
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

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
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

	for _, field := range []string{"index", "quote", "format"} {
		if err = tplValidator(attrs.MustString(field, ptrStringEmpty)); err != nil {
			return errors.Wrapf(err, "validating %s template", field)
		}
	}

	return nil
}
