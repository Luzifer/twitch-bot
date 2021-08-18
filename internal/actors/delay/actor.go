package delay

import (
	"math/rand"
	"time"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
)

func Register(args plugins.RegistrationArguments) error {
	args.RegisterActor(func() plugins.Actor { return &ActorDelay{} })

	return nil
}

type ActorDelay struct {
	Delay       time.Duration `json:"delay" yaml:"delay"`
	DelayJitter time.Duration `json:"delay_jitter" yaml:"delay_jitter"`
}

func (a ActorDelay) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule) (preventCooldown bool, err error) {
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

func (a ActorDelay) IsAsync() bool { return false }
func (a ActorDelay) Name() string  { return "delay" }
