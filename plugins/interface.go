package plugins

import (
	"github.com/go-irc/irc"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/twitch"
)

type (
	Actor interface {
		// Execute will be called after the config was read into the Actor
		Execute(c *irc.Client, m *irc.Message, r *Rule, evtData *FieldCollection, attrs *FieldCollection) (preventCooldown bool, err error)
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
		Validate(*FieldCollection) error
	}

	ActorCreationFunc func() Actor

	ActorRegistrationFunc func(name string, acf ActorCreationFunc)

	ActorDocumentationRegistrationFunc func(ActionDocumentation)

	CronRegistrationFunc func(spec string, cmd func()) (cron.EntryID, error)

	EventHandlerFunc         func(evt string, eventData *FieldCollection) error
	EventHandlerRegisterFunc func(EventHandlerFunc) error

	LoggerCreationFunc func(moduleName string) *log.Entry

	MsgFormatter func(tplString string, m *irc.Message, r *Rule, fields *FieldCollection) (string, error)

	RawMessageHandlerFunc         func(m *irc.Message) error
	RawMessageHandlerRegisterFunc func(RawMessageHandlerFunc) error

	// RegisterFunc is the type of function your plugin must expose with the name Register
	RegisterFunc func(RegistrationArguments) error

	RegistrationArguments struct {
		// FormatMessage is a method to convert templates into strings using internally known variables / configs
		FormatMessage MsgFormatter
		// GetLogger returns a sirupsen log.Entry pre-configured with the module name
		GetLogger LoggerCreationFunc
		// GetStorageManager returns an interface to access the modules storage
		GetStorageManager func() StorageManager
		// GetTwitchClient retrieves a fully configured Twitch client with initialized cache
		GetTwitchClient func() *twitch.Client
		// RegisterActor is used to register a new IRC rule-actor implementing the Actor interface
		RegisterActor ActorRegistrationFunc
		// RegisterActorDocumentation is used to register an ActorDocumentation for the config editor
		RegisterActorDocumentation ActorDocumentationRegistrationFunc
		// RegisterAPIRoute registers a new HTTP handler function including documentation
		RegisterAPIRoute HTTPRouteRegistrationFunc
		// RegisterCron is a method to register cron functions in the global cron instance
		RegisterCron CronRegistrationFunc
		// RegisterEventHandler is a method to register a handler function receiving ALL events
		RegisterEventHandler EventHandlerRegisterFunc
		// RegisterRawMessageHandler is a method to register an handler to receive ALL messages received
		RegisterRawMessageHandler RawMessageHandlerRegisterFunc
		// RegisterTemplateFunction can be used to register a new template functions
		RegisterTemplateFunction TemplateFuncRegister
		// SendMessage can be used to send a message not triggered by an event
		SendMessage SendMessageFunc
	}

	SendMessageFunc func(*irc.Message) error

	StorageManager interface {
		DeleteModuleStore(moduleUUID string) error
		GetModuleStore(moduleUUID string, storedObject StorageUnmarshaller) error
		SetModuleStore(moduleUUID string, storedObject StorageMarshaller) error
	}

	StorageMarshaller interface {
		MarshalStoredObject() ([]byte, error)
	}

	StorageUnmarshaller interface {
		UnmarshalStoredObject([]byte) error
	}

	TemplateFuncGetter   func(*irc.Message, *Rule, *FieldCollection) interface{}
	TemplateFuncRegister func(name string, fg TemplateFuncGetter)
)

func GenericTemplateFunctionGetter(f interface{}) TemplateFuncGetter {
	return func(*irc.Message, *Rule, *FieldCollection) interface{} { return f }
}
