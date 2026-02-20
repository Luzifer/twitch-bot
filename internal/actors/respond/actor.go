// Package respond contains an actor to send a message
package respond

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "respond"

var (
	formatMessage plugins.MsgFormatter
	send          plugins.SendMessageFunc
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) (err error) {
	formatMessage = args.FormatMessage
	send = args.SendMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Respond to message with a new message",
		Name:        "Respond to Message",
		Type:        "respond",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Message text to send",
				Key:             "message",
				Long:            true,
				Name:            "Message",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "Fallback message text to send if message cannot be generated",
				Key:             "fallback",
				Name:            "Fallback",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "false",
				Description:     "Send message as a native Twitch-reply to the original message",
				Key:             "as_reply",
				Name:            "As Reply",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeBool,
			},
			{
				Default:         "",
				Description:     "Send message to a different channel than the original message",
				Key:             "to_channel",
				Name:            "To Channel",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	if err = args.RegisterAPIRoute(plugins.HTTPRouteRegistrationArgs{
		Description:       "Send a message on behalf of the bot (send JSON object with `message` key)",
		HandlerFunc:       handleAPISend,
		Method:            http.MethodPost,
		Module:            actorName,
		Name:              "Send message",
		Path:              "/{channel}",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeTextPlain,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Channel to send the message to",
				Name:        "channel",
			},
		},
	}); err != nil {
		return fmt.Errorf("registering API route: %w", err)
	}

	return nil
}

type actor struct{}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	msg, err := formatMessage(attrs.MustString("message", nil), m, r, eventData)
	if err != nil {
		if !attrs.CanString("fallback") || attrs.MustString("fallback", nil) == "" {
			return false, errors.Wrap(err, "preparing response")
		}
		log.WithError(err).Error("Response message processing caused error, trying fallback")
		if msg, err = formatMessage(attrs.MustString("fallback", nil), m, r, eventData); err != nil {
			return false, errors.Wrap(err, "preparing response fallback")
		}
	}

	toChannel := plugins.DeriveChannel(m, eventData)
	if attrs.CanString("to_channel") && attrs.MustString("to_channel", nil) != "" {
		toChannel = fmt.Sprintf("#%s", strings.TrimLeft(attrs.MustString("to_channel", nil), "#"))
	}

	ircMessage := &irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			toChannel,
			msg,
		},
	}

	if attrs.MustBool("as_reply", helpers.Ptr(false)) {
		id, ok := m.Tags["id"]
		if ok {
			if ircMessage.Tags == nil {
				ircMessage.Tags = make(irc.Tags)
			}
			ircMessage.Tags["reply-parent-msg-id"] = id
		}
	}

	return false, errors.Wrap(
		send(ircMessage),
		"sending response",
	)
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "message", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "fallback", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "as_reply", Type: fieldcollection.SchemaFieldTypeBool}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "to_channel", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "message", "fallback"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

func handleAPISend(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, errors.Wrap(err, "parsing payload").Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(payload.Message) == "" {
		http.Error(w, errors.New("no message found").Error(), http.StatusBadRequest)
		return
	}

	if err := send(&irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			"#" + strings.TrimLeft(mux.Vars(r)["channel"], "#"),
			strings.TrimSpace(payload.Message),
		},
	}); err != nil {
		http.Error(w, errors.Wrap(err, "sending message").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
