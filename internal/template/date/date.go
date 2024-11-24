// Package date adds date-based helper functions for templating
package date

import (
	"fmt"
	"time"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	args.RegisterTemplateFunction("humanDateDiff", plugins.GenericTemplateFunctionGetter(NewInterval), plugins.TemplateFuncDocumentation{
		Description: `Returns a DateInterval object describing the time difference between a and b in a "human" way of counting the time (2023-02-05 -> 2024-03-05 = 1 Year, 1 Month)`,
		Syntax:      "humanDateDiff <a> <b>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:       `{{ humanDateDiff (mustToDate "2006-01-02 -0700" "2024-05-05 +0200") (mustToDate "2006-01-02 -0700" "2023-01-09 +0100") }}`,
			ExpectedOutput: "{1 3 25 23 0 0}",
		},
	})

	args.RegisterTemplateFunction("formatHumanDateDiff", plugins.GenericTemplateFunctionGetter(func(format string, d Interval) string {
		return d.Format(format)
	}), plugins.TemplateFuncDocumentation{
		Description: "Formats a DateInterval object according to the format (%Y, %M, %D, %H, %I, %S for years, months, days, hours, minutes, seconds - Lowercase letters without leading zeros)",
		Syntax:      "formatHumanDateDiff <format> <obj>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:       `{{ humanDateDiff (mustToDate "2006-01-02 -0700" "2024-05-05 +0200") (mustToDate "2006-01-02 -0700" "2023-01-09 +0100") | formatHumanDateDiff "%Y years, %M months, %D days" }}`,
			ExpectedOutput: "01 years, 03 months, 25 days",
		},
	})

	args.RegisterTemplateFunction("parseDuration", plugins.GenericTemplateFunctionGetter(func(duration string) (time.Duration, error) {
		d, err := time.ParseDuration(duration)
		if err != nil {
			return 0, fmt.Errorf("parsing duration: %w", err)
		}

		return d, nil
	}), plugins.TemplateFuncDocumentation{
		Description: `Parses a duration (i.e. 1h25m10s) into a time.Duration`,
		Syntax:      "parseDuration <duration>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:       `{{ parseDuration "1h30s" }}`,
			ExpectedOutput: "1h0m30s",
		},
	})

	args.RegisterTemplateFunction("parseDurationToSeconds", plugins.GenericTemplateFunctionGetter(func(duration string) (int64, error) {
		d, err := time.ParseDuration(duration)
		if err != nil {
			return 0, fmt.Errorf("parsing duration: %w", err)
		}

		return int64(d / time.Second), nil
	}), plugins.TemplateFuncDocumentation{
		Description: `Parses a duration (i.e. 1h25m10s) into a number of seconds`,
		Syntax:      "parseDurationToSeconds <duration>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:       `{{ parseDurationToSeconds "1h25m10s" }}`,
			ExpectedOutput: "5110",
		},
	})

	return nil
}
