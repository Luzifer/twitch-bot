package nuke

import (
	"context"
	"fmt"
	"time"
)

type (
	actionFn func(channel, match, msgid, user string) error
)

func actionBan(channel, match, _, user string) (err error) {
	if err = botTwitchClient().BanUser(
		context.Background(),
		channel,
		user,
		0,
		fmt.Sprintf("Nuke issued for %q", match),
	); err != nil {
		return fmt.Errorf("executing ban: %w", err)
	}

	return nil
}

func actionDelete(channel, _, msgid, _ string) (err error) {
	if err = botTwitchClient().DeleteMessage(
		context.Background(),
		channel,
		msgid,
	); err != nil {
		return fmt.Errorf("deleting message: %w", err)
	}

	return nil
}

func getActionTimeout(duration time.Duration) actionFn {
	return func(channel, match, _, user string) (err error) {
		if err = botTwitchClient().BanUser(
			context.Background(),
			channel,
			user,
			duration,
			fmt.Sprintf("Nuke issued for %q", match),
		); err != nil {
			return fmt.Errorf("executing timeout: %w", err)
		}

		return nil
	}
}
