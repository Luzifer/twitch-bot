package plugins

import "github.com/go-irc/irc"

type (
	Actor interface {
		// Execute will be called after the config was read into the Actor
		Execute(*irc.Client, *irc.Message, *Rule) (preventCooldown bool, err error)
		// IsAsync may return true if the Execute function is to be executed
		// in a Go routine as of long runtime. Normally it should return false
		// except in very specific cases
		IsAsync() bool
		// Name must return an unique name for the actor in order to identify
		// it in the logs for debugging purposes
		Name() string
	}

	ActorCreationFunc func() Actor

	ActorRegistrationFunc func(ActorCreationFunc)

	MsgFormatter func(tplString string, m *irc.Message, r *Rule, fields map[string]interface{}) (string, error)

	RegisterFunc func(RegistrationArguments) error

	RegistrationArguments struct {
		RegisterActor ActorRegistrationFunc
	}
)
