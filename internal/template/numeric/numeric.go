package numeric

import (
	"math"

	"github.com/Luzifer/twitch-bot/plugins"
)

func Register(args plugins.RegistrationArguments) error {
	args.RegisterTemplateFunction("add", plugins.GenericTemplateFunctionGetter(add))
	args.RegisterTemplateFunction("div", plugins.GenericTemplateFunctionGetter(div))
	args.RegisterTemplateFunction("mul", plugins.GenericTemplateFunctionGetter(mul))
	args.RegisterTemplateFunction("multiply", plugins.GenericTemplateFunctionGetter(mul)) // DEPRECATED
	args.RegisterTemplateFunction("pow", plugins.GenericTemplateFunctionGetter(math.Pow))
	args.RegisterTemplateFunction("sub", plugins.GenericTemplateFunctionGetter(sub))
	return nil
}

func add(m1, m2 float64) float64 { return m1 + m2 }
func div(m1, m2 float64) float64 { return m1 / m2 }
func mul(m1, m2 float64) float64 { return m1 * m2 }
func sub(m1, m2 float64) float64 { return m1 - m2 }
