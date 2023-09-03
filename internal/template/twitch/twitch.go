// Package twitch defines Twitch related template functions not having
// their place in any other package
package twitch

import (
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var regFn []func(plugins.RegistrationArguments)

func Register(args plugins.RegistrationArguments) error {
	for _, fn := range regFn {
		fn(args)
	}

	return nil
}
