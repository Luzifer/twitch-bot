package plugins

import (
	"github.com/go-irc/irc"
	log "github.com/sirupsen/logrus"
)

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

	LoggerCreationFunc func(moduleName string) *log.Entry

	MsgFormatter func(tplString string, m *irc.Message, r *Rule, fields map[string]interface{}) (string, error)

	// RegisterFunc is the type of function your plugin must expose with the name Register
	RegisterFunc func(RegistrationArguments) error

	RegistrationArguments struct {
		// FormatMessage is a method to convert templates into strings using internally known variables / configs
		FormatMessage MsgFormatter
		// GetLogger returns a sirupsen log.Entry pre-configured with the module name
		GetLogger LoggerCreationFunc
		// RegisterActor is used to register a new IRC rule-actor implementing the Actor interface
		RegisterActor ActorRegistrationFunc
	}
)
