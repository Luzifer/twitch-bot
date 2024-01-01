package nuke

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type (
	actionFn func(channel, match, msgid, user string) error
)

func actionBan(channel, match, _, user string) error {
	return errors.Wrap(
		botTwitchClient.BanUser(
			context.Background(),
			channel,
			user,
			0,
			fmt.Sprintf("Nuke issued for %q", match),
		),
		"executing ban",
	)
}

func actionDelete(channel, _, msgid, _ string) (err error) {
	return errors.Wrap(
		botTwitchClient.DeleteMessage(
			context.Background(),
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
				context.Background(),
				channel,
				user,
				duration,
				fmt.Sprintf("Nuke issued for %q", match),
			),
			"executing timeout",
		)
	}
}
