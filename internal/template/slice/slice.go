package slice

import (
	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/plugins"
)

func Register(args plugins.RegistrationArguments) error {
	args.RegisterTemplateFunction("inList", plugins.GenericTemplateFunctionGetter(str.StringInSlice))
	return nil
}
