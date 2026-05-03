// Package api contains helpers to interact with remote APIs in templates
package api

import "github.com/Luzifer/twitch-bot/v3/plugins"

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	args.RegisterTemplateFunction("jsonAPI", plugins.GenericTemplateFunctionGetter(jsonAPI), plugins.TemplateFuncDocumentation{
		Description: "Fetches remote URL and applies jq-like query to it returning the result as string. (Remote API needs to return status 200 within 5 seconds.)",
		Syntax:      "jsonAPI <url> <jq-like path> [fallback]",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ jsonAPI "https://api.github.com/repos/Luzifer/twitch-bot" ".owner.login" }}`,
			FakedOutput: "Luzifer",
		},
	})

	args.RegisterTemplateFunction("textAPI", plugins.GenericTemplateFunctionGetter(textAPI), plugins.TemplateFuncDocumentation{
		Description: "Fetches remote URL and returns the result as string. (Remote API needs to return status 200 within 5 seconds.)",
		Syntax:      "textAPI <url> [fallback]",
		Example: &plugins.TemplateFuncDocumentationExample{
			MatchMessage:   "!weather (.*)",
			MessageContent: "!weather Hamburg",
			Template:       `{{ textAPI (printf "https://api.scorpstuff.com/weather.php?units=metric&city=%s" (urlquery (group 1))) }}`,
			FakedOutput:    "Weather for Hamburg, DE: Few clouds with a temperature of 22 C (71.6 F). [...]",
		},
	})

	return nil
}
