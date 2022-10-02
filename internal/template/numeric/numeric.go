package numeric

import (
	"math"

	"github.com/Luzifer/twitch-bot/plugins"
)

func Register(args plugins.RegistrationArguments) error {
	args.RegisterTemplateFunction("pow", plugins.GenericTemplateFunctionGetter(math.Pow))
	return nil
}
