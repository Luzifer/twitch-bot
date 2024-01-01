// Package delay contains an actor to delay rule execution
package delay

import (
	"math/rand"
	"time"

	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "delay"

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Delay next action",
		Name:        "Delay",
		Type:        "delay",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Static delay to wait",
				Key:             "delay",
				Name:            "Delay",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeDuration,
			},
			{
				Default:         "",
				Description:     "Dynamic jitter to add to the static delay (the added extra delay will be between 0 and this value)",
				Key:             "jitter",
				Name:            "Jitter",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeDuration,
			},
		},
	})

	return nil
}

type actor struct{}

func (actor) Execute(_ *irc.Client, _ *irc.Message, _ *plugins.Rule, _ *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	var (
		ptrZeroDuration = func(v time.Duration) *time.Duration { return &v }(0)
		delay           = attrs.MustDuration("delay", ptrZeroDuration)
		jitter          = attrs.MustDuration("jitter", ptrZeroDuration)
	)

	if delay == 0 && jitter == 0 {
		return false, nil
	}

	totalDelay := delay
	if jitter > 0 {
		totalDelay += time.Duration(rand.Int63n(int64(jitter))) // #nosec: G404 // It's just time, no need for crypto/rand
	}

	time.Sleep(totalDelay)
	return false, nil
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(plugins.TemplateValidatorFunc, *plugins.FieldCollection) (err error) {
	return nil
}
