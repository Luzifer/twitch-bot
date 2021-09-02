package delay

import (
	"math/rand"
	"time"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
)

func Register(args plugins.RegistrationArguments) error {
	args.RegisterActor(func() plugins.Actor { return &actor{} })

	return nil
}

type actor struct {
	Delay       time.Duration `json:"delay" yaml:"delay"`
	DelayJitter time.Duration `json:"delay_jitter" yaml:"delay_jitter"`
}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData plugins.FieldCollection) (preventCooldown bool, err error) {
	if a.Delay == 0 && a.DelayJitter == 0 {
		return false, nil
	}

	totalDelay := a.Delay
	if a.DelayJitter > 0 {
		totalDelay += time.Duration(rand.Int63n(int64(a.DelayJitter))) // #nosec: G404 // It's just time, no need for crypto/rand
	}

	time.Sleep(totalDelay)
	return false, nil
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return "delay" }
