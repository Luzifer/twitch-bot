package twitch

import (
	"context"
	"fmt"
	"strings"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

func init() {
	regFn = append(
		regFn,
		tplTwitchDisplayName,
		tplTwitchIDForUsername,
		tplTwitchProfileImage,
		tplTwitchUserExists,
		tplTwitchUsernameForID,
	)
}

func tplTwitchDisplayName(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("displayName", plugins.GenericTemplateFunctionGetter(func(username string, v ...string) (string, error) {
		displayName, err := args.GetTwitchClient().GetDisplayNameForUser(context.Background(), strings.TrimLeft(username, "#"))
		if len(v) > 0 && (err != nil || displayName == "") {
			return v[0], nil //nolint:nilerr // Default value, no need to return error
		}

		if err != nil {
			return displayName, fmt.Errorf("getting display name: %w", err)
		}

		return displayName, nil
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
		id, err := args.GetTwitchClient().GetIDForUsername(context.Background(), username)
		if err != nil {
			return id, fmt.Errorf("getting ID for username: %w", err)
		}
		return id, nil
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
		user, err := args.GetTwitchClient().GetUserInformation(context.Background(), strings.TrimLeft(username, "#@"))
		if err != nil {
			return "", fmt.Errorf("getting user info: %w", err)
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

func tplTwitchUserExists(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("userExists", plugins.GenericTemplateFunctionGetter(func(username string) bool {
		user, err := args.GetTwitchClient().GetUserInformation(context.Background(), strings.TrimLeft(username, "#@"))
		if err != nil {
			// Well, they probably don't exist
			return false
		}

		return strings.EqualFold(username, user.Login)
	}), plugins.TemplateFuncDocumentation{
		Description: "Checks whether the given user exists",
		Syntax:      "userExists <username>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ userExists "luziferus" }}`,
			FakedOutput: "true",
		},
	})
}

func tplTwitchUsernameForID(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("usernameForID", plugins.GenericTemplateFunctionGetter(func(id string) (string, error) {
		username, err := args.GetTwitchClient().GetUsernameForID(context.Background(), id)
		if err != nil {
			return username, fmt.Errorf("getting username for ID: %w", err)
		}
		return username, nil
	}), plugins.TemplateFuncDocumentation{
		Description: "Returns the current login name of an user-id",
		Syntax:      "usernameForID <user-id>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ usernameForID "12826" }}`,
			FakedOutput: "twitch",
		},
	})
}
