package ban

import (
	"net/http"
	"regexp"

	"github.com/go-irc/irc"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v2/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v2/plugins"
)

const actorName = "ban"

var (
	botTwitchClient *twitch.Client
	formatMessage   plugins.MsgFormatter

	banChatcommandRegex = regexp.MustCompile(`^/ban +([^\s]+) +(.+)$`)
)

func Register(args plugins.RegistrationArguments) error {
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
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	args.RegisterAPIRoute(plugins.HTTPRouteRegistrationArgs{
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
				Required:    false,
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
	})

	args.RegisterMessageModFunc("/ban", handleChatCommand)

	return nil
}

type actor struct{}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	ptrStringEmpty := func(v string) *string { return &v }("")

	reason, err := formatMessage(attrs.MustString("reason", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing reason template")
	}

	return false, errors.Wrap(
		botTwitchClient.BanUser(
			plugins.DeriveChannel(m, eventData),
			plugins.DeriveUser(m, eventData),
			0,
			reason,
		),
		"executing ban",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(attrs *plugins.FieldCollection) (err error) { return nil }

func handleAPIBan(w http.ResponseWriter, r *http.Request) {
	var (
		vars    = mux.Vars(r)
		channel = vars["channel"]
		user    = vars["user"]
		reason  = r.FormValue("reason")
	)

	if err := botTwitchClient.BanUser(channel, user, 0, reason); err != nil {
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

	if err := botTwitchClient.BanUser(channel, matches[1], 0, matches[3]); err != nil {
		return errors.Wrap(err, "executing ban")
	}

	return plugins.ErrSkipSendingMessage
}
