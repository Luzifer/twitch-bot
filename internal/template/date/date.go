// Package date adds date-based helper functions for templating
package date

import (
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

	return nil
}
