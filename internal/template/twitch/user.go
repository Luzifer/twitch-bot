package twitch

import (
	"context"
	"strings"

	"github.com/Luzifer/twitch-bot/v3/plugins"
	"github.com/pkg/errors"
)

func init() {
	regFn = append(
		regFn,
		tplTwitchDisplayName,
		tplTwitchIDForUsername,
		tplTwitchProfileImage,
		tplTwitchUsernameForID,
	)
}

func tplTwitchDisplayName(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("displayName", plugins.GenericTemplateFunctionGetter(func(username string, v ...string) (string, error) {
		displayName, err := args.GetTwitchClient().GetDisplayNameForUser(strings.TrimLeft(username, "#"))
		if len(v) > 0 && (err != nil || displayName == "") {
			return v[0], nil //nolint:nilerr // Default value, no need to return error
		}

		return displayName, err
	}), plugins.TemplateFuncDocumentation{
		Description: "Returns the display name the specified user set for themselves",
		Syntax:      "displayName <username> [fallback]",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ displayName "luziferus" }} - {{ displayName "notexistinguser" "foobar" }}`,
			FakedOutput: "Luziferus - foobar",
		},
	})
}

func tplTwitchIDForUsername(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("idForUsername", plugins.GenericTemplateFunctionGetter(func(username string) (string, error) {
		return args.GetTwitchClient().GetIDForUsername(username)
	}), plugins.TemplateFuncDocumentation{
		Description: "Returns the user-id for the given username",
		Syntax:      "idForUsername <username>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ idForUsername "twitch" }}`,
			FakedOutput: "12826",
		},
	})
}

func tplTwitchProfileImage(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("profileImage", plugins.GenericTemplateFunctionGetter(func(username string) (string, error) {
		user, err := args.GetTwitchClient().GetUserInformation(strings.TrimLeft(username, "#@"))
		if err != nil {
			return "", errors.Wrap(err, "getting user info")
		}

		return user.ProfileImageURL, nil
	}), plugins.TemplateFuncDocumentation{
		Description: "Gets the URL of the given users profile image",
		Syntax:      "profileImage <username>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ profileImage .username }}`,
			FakedOutput: "https://static-cdn.jtvnw.net/jtv_user_pictures/[...].png",
		},
	})
}

func tplTwitchUsernameForID(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("usernameForID", plugins.GenericTemplateFunctionGetter(func(id string) (string, error) {
		return args.GetTwitchClient().GetUsernameForID(context.Background(), id)
	}), plugins.TemplateFuncDocumentation{
		Description: "Returns the current login name of an user-id",
		Syntax:      "usernameForID <user-id>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ usernameForID "12826" }}`,
			FakedOutput: "twitch",
		},
	})
}