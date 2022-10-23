package api

import "github.com/Luzifer/twitch-bot/v2/plugins"

func Register(args plugins.RegistrationArguments) error {
	args.RegisterTemplateFunction("jsonAPI", plugins.GenericTemplateFunctionGetter(jsonAPI))

	return nil
}
