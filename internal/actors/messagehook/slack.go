package messagehook

import (
	"net/http"
	"strings"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

type slackCompatibleActor struct {
	plugins.ActorKit
}

func (s slackCompatibleActor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	text, err := formatMessage(attrs.MustString("text", nil), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "parsing text")
	}

	return sendPayload(
		s.fixHookURL(attrs.MustString("hook_url", ptrStringEmpty)),
		map[string]string{
			"text": text,
		},
		http.StatusOK,
	)
}

func (slackCompatibleActor) IsAsync() bool { return false }

func (slackCompatibleActor) Name() string { return "slackhook" }

func (s slackCompatibleActor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
	if err = s.ValidateRequireNonEmpty(attrs, "hook_url", "text"); err != nil {
		return err
	}

	return s.ValidateRequireValidTemplate(tplValidator, attrs, "text")
}

func (slackCompatibleActor) fixHookURL(hookURL string) string {
	if strings.HasPrefix(hookURL, "https://discord.com/api/webhooks/") && !strings.HasSuffix(hookURL, "/slack") {
		hookURL = strings.Join([]string{
			strings.TrimRight(hookURL, "/"),
			"slack",
		}, "/")
	}

	return hookURL
}

func (slackCompatibleActor) register(args plugins.RegistrationArguments) {
	args.RegisterActor("slackhook", func() plugins.Actor { return &slackCompatibleActor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Sends a message to a Slack(-compatible) Web-hook",
		Name:        "Slack Message-Webhook",
		Type:        "slackhook",

		Fields: []plugins.ActionDocumentationField{
			{
				Description:     "URL to send the POST request to",
				Key:             "hook_url",
				Name:            "Hook URL",
				Optional:        false,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Description:     "Text to send to the web-hook",
				Key:             "text",
				Name:            "Message",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})
}
