package delay

import (
	"math/rand"
	"time"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
)

const actorName = "delay"

func Register(args plugins.RegistrationArguments) error {
	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	return nil
}

type actor struct{}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData plugins.FieldCollection, attrs plugins.FieldCollection) (preventCooldown bool, err error) {
	var (
		ptrZeroDuration = func(v time.Duration) *time.Duration { return &v }(0)
		delay           = attrs.MustDuration("delay", ptrZeroDuration)
		jitter          = attrs.MustDuration("delay_jitter", ptrZeroDuration)
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

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(attrs plugins.FieldCollection) (err error) { return nil }
