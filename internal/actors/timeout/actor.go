package timeout

import (
	"fmt"
	"time"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

const actorName = "timeout"

func Register(args plugins.RegistrationArguments) error {
	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	return nil
}

type actor struct {
	Timeout *time.Duration `json:"timeout" yaml:"timeout"`
}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData plugins.FieldCollection) (preventCooldown bool, err error) {
	if a.Timeout == nil {
		return false, nil
	}

	return false, errors.Wrap(
		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				plugins.DeriveChannel(m, eventData),
				fmt.Sprintf("/timeout %s %d", plugins.DeriveUser(m, eventData), fixDurationValue(*a.Timeout)/time.Second),
			},
		}),
		"sending timeout",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func fixDurationValue(d time.Duration) time.Duration {
	if d >= time.Second {
		return d
	}

	return d * time.Second
}
