package api

import "github.com/Luzifer/twitch-bot/v3/plugins"

func Register(args plugins.RegistrationArguments) error {
	args.RegisterTemplateFunction("jsonAPI", plugins.GenericTemplateFunctionGetter(jsonAPI))
	args.RegisterTemplateFunction("textAPI", plugins.GenericTemplateFunctionGetter(textAPI))

	return nil
}
