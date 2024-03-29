// Package ban contains actors to ban/unban users in a channel
package ban

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "ban"

var (
	botTwitchClient *twitch.Client
	formatMessage   plugins.MsgFormatter

	banChatcommandRegex = regexp.MustCompile(`^/ban +([^\s]+) +(.+)$`)
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) (err error) {
	botTwitchClient = args.GetTwitchClient()
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

type actor struct{}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	ptrStringEmpty := func(v string) *string { return &v }("")

	reason, err := formatMessage(attrs.MustString("reason", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing reason template")
	}

	return false, errors.Wrap(
		botTwitchClient.BanUser(
			context.Background(),
			plugins.DeriveChannel(m, eventData),
			plugins.DeriveUser(m, eventData),
			0,
			reason,
		),
		"executing ban",
	)
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
	reasonTemplate, err := attrs.String("reason")
	if err != nil || reasonTemplate == "" {
		return errors.New("reason must be non-empty string")
	}

	if err = tplValidator(reasonTemplate); err != nil {
		return errors.Wrap(err, "validating reason template")
	}

	return nil
}

func handleAPIBan(w http.ResponseWriter, r *http.Request) {
	var (
		vars    = mux.Vars(r)
		channel = vars["channel"]
		user    = vars["user"]
		reason  = r.FormValue("reason")
	)

	if err := botTwitchClient.BanUser(r.Context(), channel, user, 0, reason); err != nil {
		http.Error(w, errors.Wrap(err, "issuing ban").Error(), http.StatusInternalServerError)
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

	if err := botTwitchClient.BanUser(context.Background(), channel, matches[1], 0, matches[2]); err != nil {
		return errors.Wrap(err, "executing ban")
	}

	return plugins.ErrSkipSendingMessage
}
