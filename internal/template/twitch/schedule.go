package twitch

import (
	"context"
	"math"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
	"github.com/pkg/errors"
)

func init() {
	regFn = append(
		regFn,
		tplTwitchScheduleSegments,
	)
}

func tplTwitchScheduleSegments(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("scheduleSegments", plugins.GenericTemplateFunctionGetter(func(channel string, n ...int) ([]twitch.ChannelStreamScheduleSegment, error) {
		schedule, err := args.GetTwitchClient().GetChannelStreamSchedule(context.Background(), channel)
		if err != nil {
			return nil, errors.Wrap(err, "getting schedule")
		}

		if len(n) > 0 {
			return schedule.Segments[:int(math.Min(float64(n[0]), float64(len(schedule.Segments))))], nil
		}

		return schedule.Segments, nil
	}), plugins.TemplateFuncDocumentation{
		Description: "Returns the next n segments in the channels schedule. If n is not given, returns all known segments.",
		Syntax:      "scheduleSegments <channel> [n]",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ $seg := scheduleSegments "luziferus" 1 | first }}Next Stream: {{ $seg.Title }} @	{{ dateInZone "2006-01-02 15:04" $seg.StartTime "Europe/Berlin" }}`,
			FakedOutput: "Next Stream: Little Nightmares @ 2023-11-05 18:00",
		},
	})
}
