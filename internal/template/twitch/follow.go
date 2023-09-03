package twitch

import (
	"time"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
	"github.com/pkg/errors"
)

func init() {
	regFn = append(
		regFn,
		tplTwitchDoesFollow,
		tplTwitchDoesFollowLongerThan,
		tplTwitchFollowAge,
		tplTwitchFollowDate,
	)
}

func tplTwitchDoesFollowLongerThan(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("doesFollowLongerThan", plugins.GenericTemplateFunctionGetter(func(from, to string, t any) (bool, error) {
		var (
			age time.Duration
			err error
		)

		switch v := t.(type) {
		case int64:
			age = time.Duration(v) * time.Second

		case string:
			if age, err = time.ParseDuration(v); err != nil {
				return false, errors.Wrap(err, "parsing duration")
			}

		default:
			return false, errors.Errorf("unexpected input for duration %t", t)
		}

		fd, err := args.GetTwitchClient().GetFollowDate(from, to)
		switch {
		case err == nil:
			return time.Since(fd) > age, nil

		case errors.Is(err, twitch.ErrUserDoesNotFollow):
			return false, nil

		default:
			return false, errors.Wrap(err, "getting follow date")
		}
	}), plugins.TemplateFuncDocumentation{
		Description: "Returns whether `from` follows `to` for more than `duration`",
		Syntax:      "doesFollowLongerThan <from> <to> <duration>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ doesFollowLongerThan "tezrian" "luziferus" "168h" }}`,
			FakedOutput: "true",
		},
	})
}

func tplTwitchDoesFollow(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("doesFollow", plugins.GenericTemplateFunctionGetter(func(from, to string) (bool, error) {
		_, err := args.GetTwitchClient().GetFollowDate(from, to)
		switch {
		case err == nil:
			return true, nil

		case errors.Is(err, twitch.ErrUserDoesNotFollow):
			return false, nil

		default:
			return false, errors.Wrap(err, "getting follow date")
		}
	}), plugins.TemplateFuncDocumentation{
		Description: "Returns whether `from` follows `to`",
		Syntax:      "doesFollow <from> <to>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ doesFollow "tezrian" "luziferus" }}`,
			FakedOutput: "true",
		},
	})
}

func tplTwitchFollowAge(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("followAge", plugins.GenericTemplateFunctionGetter(func(from, to string) (time.Duration, error) {
		since, err := args.GetTwitchClient().GetFollowDate(from, to)
		return time.Since(since), errors.Wrap(err, "getting follow date")
	}), plugins.TemplateFuncDocumentation{
		Description: "Looks up when `from` followed `to` and returns the duration between then and now",
		Syntax:      "followAge <from> <to>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ followAge "tezrian" "luziferus" }}`,
			FakedOutput: "15004h14m59.116620989s",
		},
	})
}

func tplTwitchFollowDate(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("followDate", plugins.GenericTemplateFunctionGetter(func(from, to string) (time.Time, error) {
		return args.GetTwitchClient().GetFollowDate(from, to)
	}), plugins.TemplateFuncDocumentation{
		Description: "Looks up when `from` followed `to`",
		Syntax:      "followDate <from> <to>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ followDate "tezrian" "luziferus" }}`,
			FakedOutput: "2021-04-10 16:07:07 +0000 UTC",
		},
	})
}
