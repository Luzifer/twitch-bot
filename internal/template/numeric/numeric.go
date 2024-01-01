// Package numeric contains helpers for numeric manipulation
package numeric

import (
	"math"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	args.RegisterTemplateFunction("pow", plugins.GenericTemplateFunctionGetter(math.Pow), plugins.TemplateFuncDocumentation{
		Description: "Returns float from calculation: `float1 ** float2`",
		Syntax:      "pow <float1> <float2>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:       `{{ printf "%.0f" (pow 10 4) }}`,
			ExpectedOutput: "10000",
		},
	})

	return nil
}
