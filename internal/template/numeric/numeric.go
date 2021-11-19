package numeric

import "github.com/Luzifer/twitch-bot/plugins"

func Register(args plugins.RegistrationArguments) error {
	args.RegisterTemplateFunction("multiply", plugins.GenericTemplateFunctionGetter(multiply))
	return nil
}

func multiply(m1, m2 float64) float64 { return m1 * m2 }
