package nuke

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type (
	actionFn func(channel, match, msgid, user string) error
)

func actionBan(channel, match, msgid, user string) error {
	return errors.Wrap(
		botTwitchClient.BanUser(
			channel,
			user,
			0,
			fmt.Sprintf("Nuke issued for %q", match),
		),
		"executing ban",
	)
}

func actionDelete(channel, match, msgid, user string) (err error) {
	return errors.Wrap(
		botTwitchClient.DeleteMessage(
			channel,
			msgid,
		),
		"deleting message",
	)
}

func getActionTimeout(duration time.Duration) actionFn {
	return func(channel, match, msgid, user string) error {
		return errors.Wrap(
			botTwitchClient.BanUser(
				channel,
				user,
				duration,
				fmt.Sprintf("Nuke issued for %q", match),
			),
			"executing timeout",
		)
	}
}
