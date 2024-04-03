package plugins

import (
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"
	"gorm.io/gorm"

	"github.com/Luzifer/go_helpers/v2/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
)

type (
	// Actor defines an interface to implement in the plugin for actors
	Actor interface {
		// Execute will be called after the config was read into the Actor
		Execute(c *irc.Client, m *irc.Message, r *Rule, evtData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error)
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
		Validate(TemplateValidatorFunc, *fieldcollection.FieldCollection) error
	}

	// ActorCreationFunc is a function to return a new instance of the
	// plugins actor
	ActorCreationFunc func() Actor

	// ActorRegistrationFunc is passed from the bot to the plugins
	// RegisterFunc to register a new actor in the bot
	ActorRegistrationFunc func(name string, acf ActorCreationFunc)

	// ActorDocumentationRegistrationFunc is passed from the bot to the
	// plugins RegisterFunc to register a new actor documentation
	ActorDocumentationRegistrationFunc func(ActionDocumentation)

	// ChannelPermissionCheckFunc is available to check whether the bot
	// has stored scopes / permissions for the given channel. All given
	// scopes need to be available to return true.
	ChannelPermissionCheckFunc func(channel string, scopes ...string) (bool, error)
	// ChannelAnyPermissionCheckFunc is available to check whether the bot
	// has stored scopes / permissions for the given channel. Any of the
	// given scopes need to be available to return true.
	ChannelAnyPermissionCheckFunc func(channel string, scopes ...string) (bool, error)

	// CronRegistrationFunc is passed from the bot to the
	// plugins RegisterFunc to register a new cron function in the
	// internal cron scheduler
	CronRegistrationFunc func(spec string, cmd func()) (cron.EntryID, error)

	// DatabaseCopyFunc defines the function type the plugin must
	// implement and register to enable the bot to replicate its database
	// stored content into a new database
	DatabaseCopyFunc func(src, target *gorm.DB) error

	// EventHandlerFunc defines the type of function required to listen
	// for events
	EventHandlerFunc func(evt string, eventData *fieldcollection.FieldCollection) error
	// EventHandlerRegisterFunc is passed from the bot to the
	// plugins RegisterFunc to register a new event handler function
	// which is then fed with all events occurring in the bot
	EventHandlerRegisterFunc func(EventHandlerFunc) error

	// LoggerCreationFunc is passed from the bot to the
	// plugins RegisterFunc to retrieve a pre-configured logrus.Entry
	// scoped for the given module name
	LoggerCreationFunc func(moduleName string) *logrus.Entry

	// ModuleConfigGetterFunc is passed from the bot to the
	// plugins RegisterFunc to fetch module generic or channel specific
	// configuration from the module configuration
	ModuleConfigGetterFunc func(module, channel string) *fieldcollection.FieldCollection

	// MsgFormatter is passed from the bot to the
	// plugins RegisterFunc to format messages using all registered and
	// available template functions
	MsgFormatter func(tplString string, m *irc.Message, r *Rule, fields *fieldcollection.FieldCollection) (string, error)

	// MsgModificationFunc can be used to modify messages between the
	// plugins generating them and the bot sending them to the Twitch
	// servers
	MsgModificationFunc func(*irc.Message) error
	// MsgModificationRegistrationFunc is passed from the bot to the
	// plugins RegisterFunc to register a new MsgModificationFunc for
	// the given prefix
	MsgModificationRegistrationFunc func(linePrefix string, modFn MsgModificationFunc)

	// RawMessageHandlerFunc is the type of function to implement in
	// your plugin in order to process raw-messages from the IRC
	// connection
	RawMessageHandlerFunc func(m *irc.Message) error
	// RawMessageHandlerRegisterFunc is passed from the bot to the
	// plugins RegisterFunc to register a new RawMessageHandlerFunc
	RawMessageHandlerRegisterFunc func(RawMessageHandlerFunc) error

	// RegisterFunc is the type of function your plugin must expose with the name Register
	RegisterFunc func(RegistrationArguments) error

	// RegistrationArguments is the object your RegisterFunc will receive
	// and can use to interact with the bot instance
	RegistrationArguments struct {
		// CreateEvent allows to create an event handed out to all modules to handle
		CreateEvent EventHandlerFunc
		// FormatMessage is a method to convert templates into strings using internally known variables / configs
		FormatMessage MsgFormatter
		// FrontendNotify is a way to send a notification to the frontend
		FrontendNotify func(string)
		// GetBaseURL returns the configured BaseURL for the bot
		GetBaseURL func() string
		// GetDatabaseConnector returns an active database.Connector to access the backend storage database
		GetDatabaseConnector func() database.Connector
		// GetLogger returns a sirupsen log.Entry pre-configured with the module name
		GetLogger LoggerCreationFunc
		// GetModuleConfigForChannel returns the module configuration for the given channel if available
		GetModuleConfigForChannel ModuleConfigGetterFunc
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
		// RegisterCopyDatabaseFunc registers a DatabaseCopyFunc for the
		// database migration tool. Modules not registering such a func
		// will not be copied over when migrating to another database.
		RegisterCopyDatabaseFunc func(name string, fn DatabaseCopyFunc)
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

	// SendMessageFunc is available through the RegistrationArguments
	// and MUST be used to send messages to the Twitch servers
	SendMessageFunc func(*irc.Message) error

	// TemplateFuncGetter is the type of function to implement in the
	// plugin to create a new template function on request of the bot
	TemplateFuncGetter func(*irc.Message, *Rule, *fieldcollection.FieldCollection) any
	// TemplateFuncRegister is passed from the bot to the
	// plugins RegisterFunc to register a new TemplateFuncGetter
	TemplateFuncRegister func(name string, fg TemplateFuncGetter, doc ...TemplateFuncDocumentation)

	// TemplateValidatorFunc is passed from the bot to the
	// plugins RegisterFunc to validate templates considering all
	// registered template functions
	TemplateValidatorFunc func(raw string) error

	// ValidateTokenFunc is passed from the bot to the
	// plugins RegisterFunc to validate tokens and their access
	// permissions
	ValidateTokenFunc func(token string, modules ...string) error
)

// ErrSkipSendingMessage should be returned by a MsgModificationFunc
// to prevent the message to be sent to the Twitch servers
var ErrSkipSendingMessage = errors.New("skip sending message")

// GenericTemplateFunctionGetter wraps a generic template function not
// requiring access to the irc.Message, Rule or FieldCollection to
// satisfy the TemplateFuncGetter interface
func GenericTemplateFunctionGetter(f any) TemplateFuncGetter {
	return func(*irc.Message, *Rule, *fieldcollection.FieldCollection) any { return f }
}
