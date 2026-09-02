package overlays

import (
	"fmt"
	"time"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

func execCleanupCron(args plugins.RegistrationArguments) error {
	channels, err := getChannelsWithStoredEvents(args.GetDatabaseConnector())
	if err != nil {
		return fmt.Errorf("querying channels: %w", err)
	}

	for _, channel := range channels {
		conf := args.GetModuleConfigForChannel("overlays", channel)
		if !conf.CanDuration("event_retention") {
			// No default or channel configuration: Don't clean.
			continue
		}

		eventRetention := conf.MustDuration("event_retention", nil)
		if eventRetention <= 0 {
			// No valid time given
			continue
		}

		deleteOlderThan := time.Now().Add(-eventRetention)
		if err = deleteEventsOlderThan(args.GetDatabaseConnector(), channel, deleteOlderThan); err != nil {
			return err
		}
	}

	return nil
}
