package plugins

import (
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
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
		Validate(TemplateValidatorFunc, *FieldCollection) error
	}

	ActorCreationFunc func() Actor

	ActorRegistrationFunc func(name string, acf ActorCreationFunc)

	ActorDocumentationRegistrationFunc func(ActionDocumentation)

	ChannelPermissionCheckFunc    func(channel string, scopes ...string) (bool, error)
	ChannelAnyPermissionCheckFunc func(channel string, scopes ...string) (bool, error)

	CronRegistrationFunc func(spec string, cmd func()) (cron.EntryID, error)

	EventHandlerFunc         func(evt string, eventData *FieldCollection) error
	EventHandlerRegisterFunc func(EventHandlerFunc) error

	LoggerCreationFunc func(moduleName string) *log.Entry

	MsgFormatter func(tplString string, m *irc.Message, r *Rule, fields *FieldCollection) (string, error)

	MsgModificationFunc             func(*irc.Message) error
	MsgModificationRegistrationFunc func(linePrefix string, modFn MsgModificationFunc)

	RawMessageHandlerFunc         func(m *irc.Message) error
	RawMessageHandlerRegisterFunc func(RawMessageHandlerFunc) error

	// RegisterFunc is the type of function your plugin must expose with the name Register
	RegisterFunc func(RegistrationArguments) error

	RegistrationArguments struct {
		// CreateEvent allows to create an event handed out to all modules to handle
		CreateEvent EventHandlerFunc
		// FormatMessage is a method to convert templates into strings using internally known variables / configs
		FormatMessage MsgFormatter
		// FrontendNotify is a way to send a notification to the frontend
		FrontendNotify func(string)
		// GetDatabaseConnector returns an active database.Connector to access the backend storage database
		GetDatabaseConnector func() database.Connector
		// GetLogger returns a sirupsen log.Entry pre-configured with the module name
		GetLogger LoggerCreationFunc
		// GetTwitchClient retrieves a fully configured Twitch client with initialized cache
		GetTwitchClient func() *twitch.Client
		// GetTwitchClientForChannel retrieves a fully configured Twitch client with initialized cache for extended permission channels
		GetTwitchClientForChannel func(string) (*twitch.Client, error)
		// HasAnyPermissionForChannel checks whether ANY of the given permissions were granted for the given channel
		HasAnyPermissionForChannel ChannelAnyPermissionCheckFunc
		// HasPermissionForChannel checks whether ALL of the given permissions were granted for the given channel
		HasPermissionForChannel ChannelPermissionCheckFunc
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
		// RegisterMessageModFunc is a method to register a handler to modify / react on messages
		RegisterMessageModFunc MsgModificationRegistrationFunc
		// RegisterRawMessageHandler is a method to register an handler to receive ALL messages received
		RegisterRawMessageHandler RawMessageHandlerRegisterFunc
		// RegisterTemplateFunction can be used to register a new template functions
		RegisterTemplateFunction TemplateFuncRegister
		// SendMessage can be used to send a message not triggered by an event
		SendMessage SendMessageFunc
		// ValidateToken offers a way to validate a token and determine whether it has permissions on a given module
		ValidateToken ValidateTokenFunc
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

	TemplateValidatorFunc func(raw string) error

	ValidateTokenFunc func(token string, modules ...string) error
)

var ErrSkipSendingMessage = errors.New("skip sending message")

func GenericTemplateFunctionGetter(f interface{}) TemplateFuncGetter {
	return func(*irc.Message, *Rule, *FieldCollection) interface{} { return f }
}
