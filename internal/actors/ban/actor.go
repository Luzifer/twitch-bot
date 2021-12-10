package ban

import (
	"net/http"
	"strings"

	"github.com/go-irc/irc"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/plugins"
)

const actorName = "ban"

var (
	formatMessage plugins.MsgFormatter
	send          plugins.SendMessageFunc
)

func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage
	send = args.SendMessage

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
		HandlerFunc: handleAPIBan,
		Method:      http.MethodPost,
		Module:      "ban",
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

	return nil
}

type actor struct{}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	ptrStringEmpty := func(v string) *string { return &v }("")

	reason, err := formatMessage(attrs.MustString("reason", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing reason template")
	}

	return a.execBan(
		plugins.DeriveChannel(m, eventData),
		plugins.DeriveUser(m, eventData),
		reason,
	)
}

func (actor) execBan(channel, user, reason string) (bool, error) {
	cmd := []string{
		"/ban",
		user,
	}

	if reason != "" {
		cmd = append(cmd, reason)
	}

	return false, errors.Wrap(
		send(&irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				"#" + strings.TrimLeft(channel, "#"),
				strings.Join(cmd, " "),
			},
		}),
		"sending ban",
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

	if _, err := (actor{}).execBan(channel, user, reason); err != nil {
		http.Error(w, errors.Wrap(err, "issuing ban").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
