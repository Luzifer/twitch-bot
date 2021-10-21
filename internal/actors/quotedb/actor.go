package quotedb

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"sync"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

const (
	actorName  = "quotedb"
	moduleUUID = "917c83ee-ed40-41e4-a558-1c2e59fdf1f5"
)

var (
	formatMessage plugins.MsgFormatter
	store         plugins.StorageManager
	storedObject  = newStorage()

	ptrStringEmpty     = func(v string) *string { return &v }("")
	ptrStringOutFormat = func(v string) *string { return &v }("Quote #{{ .index }}: {{ .quote }}")
	ptrStringZero      = func(v string) *string { return &v }("0")
)

func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage
	store = args.GetStorageManager()

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

	args.RegisterTemplateFunction("lastQuoteIndex", func(m *irc.Message, r *plugins.Rule, fields plugins.FieldCollection) interface{} {
		return func() int {
			return storedObject.GetMaxQuoteIdx(plugins.DeriveChannel(m, nil))
		}
	})

	return errors.Wrap(
		store.GetModuleStore(moduleUUID, storedObject),
		"loading module storage",
	)
}

type (
	actor struct{}

	storage struct {
		ChannelQuotes map[string][]string `json:"channel_quotes"`

		lock sync.RWMutex
	}
)

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData plugins.FieldCollection, attrs plugins.FieldCollection) (preventCooldown bool, err error) {
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

		storedObject.AddQuote(plugins.DeriveChannel(m, eventData), quote)
		return false, errors.Wrap(
			store.SetModuleStore(moduleUUID, storedObject),
			"storing quote database",
		)

	case "del":
		storedObject.DelQuote(plugins.DeriveChannel(m, eventData), index)
		return false, errors.Wrap(
			store.SetModuleStore(moduleUUID, storedObject),
			"storing quote database",
		)

	case "get":
		idx, quote := storedObject.GetQuote(plugins.DeriveChannel(m, eventData), index)

		if idx == 0 {
			// No quote was found for the given idx
			return false, nil
		}

		fields := make(plugins.FieldCollection)
		for k, v := range eventData {
			fields[k] = v
		}
		fields["index"] = idx
		fields["quote"] = quote

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

func (a actor) Validate(attrs plugins.FieldCollection) (err error) {
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

// Storage

func newStorage() *storage {
	return &storage{
		ChannelQuotes: make(map[string][]string),
	}
}

func (s *storage) AddQuote(channel, quote string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.ChannelQuotes[channel] = append(s.ChannelQuotes[channel], quote)
}

func (s *storage) DelQuote(channel string, quote int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	var quotes []string
	for i, q := range s.ChannelQuotes[channel] {
		if i == quote {
			continue
		}
		quotes = append(quotes, q)
	}

	s.ChannelQuotes[channel] = quotes
}

func (s *storage) GetChannelQuotes(channel string) []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var out []string
	out = append(out, s.ChannelQuotes[channel]...)
	return out
}

func (s *storage) GetMaxQuoteIdx(channel string) int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.ChannelQuotes[channel])
}

func (s *storage) GetQuote(channel string, quote int) (int, string) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if quote == 0 {
		quote = rand.Intn(len(s.ChannelQuotes[channel])) + 1 // #nosec G404 // no need for cryptographic safety
	}

	if quote > len(s.ChannelQuotes[channel]) {
		return 0, ""
	}

	return quote, s.ChannelQuotes[channel][quote-1]
}

func (s *storage) SetQuotes(channel string, quotes []string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.ChannelQuotes[channel] = quotes
}

func (s *storage) UpdateQuote(channel string, idx int, quote string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	var quotes []string
	for i := range s.ChannelQuotes[channel] {
		if i == idx {
			quotes = append(quotes, quote)
			continue
		}

		quotes = append(quotes, s.ChannelQuotes[channel][i])
	}

	s.ChannelQuotes[channel] = quotes
}

// Implement marshaller interfaces
func (s *storage) MarshalStoredObject() ([]byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return json.Marshal(s)
}

func (s *storage) UnmarshalStoredObject(data []byte) error {
	if data == nil {
		// No data set yet, don't try to unmarshal
		return nil
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	return json.Unmarshal(data, s)
}
