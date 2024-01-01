// Package slice contains slice manipulation helpers
package slice

import (
	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	args.RegisterTemplateFunction("inList", plugins.GenericTemplateFunctionGetter(func(search string, list ...string) bool {
		return str.StringInSlice(search, list)
	}), plugins.TemplateFuncDocumentation{
		Description: "Tests whether a string is in a given list of strings (for conditional templates).",
		Syntax:      "inList <search> <...string>",
		Example: &plugins.TemplateFuncDocumentationExample{
			MatchMessage:   "!command (.*)",
			MessageContent: "!command foo",
			Template:       `{{ inList (group 1) "foo" "bar" }}`,
			ExpectedOutput: "true",
		},
	})
	return nil
}
