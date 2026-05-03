// Package ban contains actors to ban/unban users in a channel
package ban

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/gorilla/mux"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "ban"

type actor struct{}

var (
	botTwitchClient func() *twitch.Client
	formatMessage   plugins.MsgFormatter

	banChatcommandRegex = regexp.MustCompile(`^/ban +([^\s]+) +(.+)$`)
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) (err error) {
	botTwitchClient = args.GetTwitchClient
	formatMessage = args.FormatMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Ban user from chat",
		Name:        "Ban User",
		Type:        "ban",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Reason why the user was banned",
				Key:             "reason",
				Name:            "Reason",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	if err = args.RegisterAPIRoute(plugins.HTTPRouteRegistrationArgs{
		Description: "Executes a ban of an user in the specified channel",
		HandlerFunc: handleAPIBan,
		Method:      http.MethodPost,
		Module:      "ban",
		Name:        "Ban User",
		Path:        "/{channel}/{user}",
		QueryParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Reason to add to the ban",
				Name:        "reason",
				Required:    true,
				Type:        "string",
			},
		},
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeTextPlain,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Channel to ban the user in",
				Name:        "channel",
			},
			{
				Description: "User to ban",
				Name:        "user",
			},
		},
	}); err != nil {
		return fmt.Errorf("registering API route: %w", err)
	}

	args.RegisterMessageModFunc("/ban", handleChatCommand)

	return nil
}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	ptrStringEmpty := func(v string) *string { return &v }("")

	reason, err := formatMessage(attrs.MustString("reason", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, fmt.Errorf("executing reason template: %w", err)
	}

	if err = botTwitchClient().BanUser(
		context.Background(),
		plugins.DeriveChannel(m, eventData),
		plugins.DeriveUser(m, eventData),
		0,
		reason,
	); err != nil {
		return false, fmt.Errorf("executing ban: %w", err)
	}

	return false, nil
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "reason", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "reason"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

func handleAPIBan(w http.ResponseWriter, r *http.Request) {
	var (
		vars    = mux.Vars(r)
		channel = vars["channel"]
		user    = vars["user"]
		reason  = r.FormValue("reason") //#nosec:G120 // Request body size is limited by API route registration middleware
	)

	if err := botTwitchClient().BanUser(r.Context(), channel, user, 0, reason); err != nil {
		http.Error(w, fmt.Errorf("issuing ban: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleChatCommand(m *irc.Message) error {
	channel := plugins.DeriveChannel(m, nil)

	matches := banChatcommandRegex.FindStringSubmatch(m.Trailing())
	if matches == nil {
		return errors.New("ban message does not match required format")
	}

	if err := botTwitchClient().BanUser(context.Background(), channel, matches[1], 0, matches[2]); err != nil {
		return fmt.Errorf("executing ban: %w", err)
	}

	return plugins.ErrSkipSendingMessage
}
