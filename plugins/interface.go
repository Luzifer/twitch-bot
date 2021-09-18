package plugins

import (
	"github.com/Luzifer/twitch-bot/twitch"
	"github.com/go-irc/irc"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

type (
	Actor interface {
		// Execute will be called after the config was read into the Actor
		Execute(c *irc.Client, m *irc.Message, r *Rule, evtData FieldCollection, attrs FieldCollection) (preventCooldown bool, err error)
		// IsAsync may return true if the Execute function is to be executed
		// in a Go routine as of long runtime. Normally it should return false
		// except in very specific cases
		IsAsync() bool
		// Name must return an unique name for the actor in order to identify
		// it in the logs for debugging purposes
		Name() string
		// Validate will be called to validate the loaded configuration. It should
		// return an error if required keys are missing from the AttributeStore
		// or if keys contain broken configs
		Validate(FieldCollection) error
	}

	ActorCreationFunc func() Actor

	ActorRegistrationFunc func(name string, acf ActorCreationFunc)

	CronRegistrationFunc func(spec string, cmd func()) (cron.EntryID, error)

	LoggerCreationFunc func(moduleName string) *log.Entry

	MsgFormatter func(tplString string, m *irc.Message, r *Rule, fields FieldCollection) (string, error)

	RawMessageHandlerFunc         func(m *irc.Message) error
	RawMessageHandlerRegisterFunc func(RawMessageHandlerFunc) error

	// RegisterFunc is the type of function your plugin must expose with the name Register
	RegisterFunc func(RegistrationArguments) error

	RegistrationArguments struct {
		// FormatMessage is a method to convert templates into strings using internally known variables / configs
		FormatMessage MsgFormatter
		// GetLogger returns a sirupsen log.Entry pre-configured with the module name
		GetLogger LoggerCreationFunc
		// GetTwitchClient retrieves a fully configured Twitch client with initialized cache
		GetTwitchClient func() *twitch.Client
		// RegisterActor is used to register a new IRC rule-actor implementing the Actor interface
		RegisterActor ActorRegistrationFunc
		// RegisterAPIRoute registers a new HTTP handler function including documentation
		RegisterAPIRoute HTTPRouteRegistrationFunc
		// RegisterCron is a method to register cron functions in the global cron instance
		RegisterCron CronRegistrationFunc
		// RegisterRawMessageHandler is a method to register an handler to receive ALL messages received
		RegisterRawMessageHandler RawMessageHandlerRegisterFunc
		// RegisterTemplateFunction can be used to register a new template functions
		RegisterTemplateFunction TemplateFuncRegister
		// SendMessage can be used to send a message not triggered by an event
		SendMessage SendMessageFunc
	}

	SendMessageFunc func(*irc.Message) error

	TemplateFuncGetter   func(*irc.Message, *Rule, map[string]interface{}) interface{}
	TemplateFuncRegister func(name string, fg TemplateFuncGetter)
)

func GenericTemplateFunctionGetter(f interface{}) TemplateFuncGetter {
	return func(*irc.Message, *Rule, map[string]interface{}) interface{} { return f }
}
