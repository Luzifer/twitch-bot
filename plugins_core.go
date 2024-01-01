package main

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/v2/backoff"
	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/announce"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/ban"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/clip"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/clipdetector"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/commercial"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/counter"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/delay"
	deleteactor "github.com/Luzifer/twitch-bot/v3/internal/actors/delete"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/eventmod"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/filesay"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/linkdetector"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/linkprotect"
	logActor "github.com/Luzifer/twitch-bot/v3/internal/actors/log"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/messagehook"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/modchannel"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/nuke"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/punish"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/quotedb"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/raw"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/respond"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/shield"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/shoutout"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/stopexec"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/timeout"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/variables"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/vip"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/whisper"
	"github.com/Luzifer/twitch-bot/v3/internal/apimodules/customevent"
	"github.com/Luzifer/twitch-bot/v3/internal/apimodules/msgformat"
	"github.com/Luzifer/twitch-bot/v3/internal/apimodules/overlays"
	"github.com/Luzifer/twitch-bot/v3/internal/apimodules/raffle"
	"github.com/Luzifer/twitch-bot/v3/internal/service/access"
	"github.com/Luzifer/twitch-bot/v3/internal/template/api"
	"github.com/Luzifer/twitch-bot/v3/internal/template/numeric"
	"github.com/Luzifer/twitch-bot/v3/internal/template/random"
	"github.com/Luzifer/twitch-bot/v3/internal/template/slice"
	"github.com/Luzifer/twitch-bot/v3/internal/template/strings"
	"github.com/Luzifer/twitch-bot/v3/internal/template/subscriber"
	twitchFns "github.com/Luzifer/twitch-bot/v3/internal/template/twitch"
	"github.com/Luzifer/twitch-bot/v3/internal/template/userstate"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const ircHandleWaitRetries = 10

var (
	corePluginRegistrations = []plugins.RegisterFunc{
		// Actors
		announce.Register,
		ban.Register,
		clip.Register,
		clipdetector.Register,
		commercial.Register,
		counter.Register,
		delay.Register,
		deleteactor.Register,
		eventmod.Register,
		filesay.Register,
		linkdetector.Register,
		linkprotect.Register,
		logActor.Register,
		messagehook.Register,
		modchannel.Register,
		nuke.Register,
		punish.Register,
		quotedb.Register,
		raw.Register,
		respond.Register,
		shield.Register,
		shoutout.Register,
		stopexec.Register,
		timeout.Register,
		variables.Register,
		vip.Register,
		whisper.Register,

		// Template functions
		api.Register,
		numeric.Register,
		random.Register,
		slice.Register,
		strings.Register,
		subscriber.Register,
		twitchFns.Register,
		userstate.Register,

		// API-only modules
		customevent.Register,
		msgformat.Register,
		overlays.Register,
		raffle.Register,
	}
	knownModules []string
)

func initCorePlugins() error {
	args := getRegistrationArguments()
	for idx, rf := range corePluginRegistrations {
		if err := rf(args); err != nil {
			return errors.Wrapf(err, "registering core plugin %d", idx)
		}
	}
	return nil
}

func registerRoute(route plugins.HTTPRouteRegistrationArgs) error {
	r := router.
		PathPrefix(fmt.Sprintf("/%s/", route.Module)).
		Subrouter()

	if !str.StringInSlice(route.Module, knownModules) {
		knownModules = append(knownModules, route.Module)
	}

	var hdl http.Handler = route.HandlerFunc
	switch {
	case route.RequiresEditorsAuth:
		hdl = writeAuthMiddleware(hdl, moduleConfigEditor)
	case route.RequiresWriteAuth:
		hdl = writeAuthMiddleware(hdl, route.Module)
	}

	if route.IsPrefix {
		r.PathPrefix(route.Path).
			Handler(hdl).
			Methods(route.Method)
	} else {
		r.Handle(route.Path, hdl).
			Methods(route.Method)
	}

	if !route.SkipDocumentation {
		return errors.Wrap(registerSwaggerRoute(route), "registering documentation")
	}

	return nil
}

func getRegistrationArguments() plugins.RegistrationArguments {
	return plugins.RegistrationArguments{
		FormatMessage:              formatMessage,
		FrontendNotify:             func(mt string) { frontendNotifyHooks.Ping(mt) },
		GetDatabaseConnector:       func() database.Connector { return db },
		GetLogger:                  func(moduleName string) *log.Entry { return log.WithField("module", moduleName) },
		GetTwitchClient:            func() *twitch.Client { return twitchClient },
		HasAnyPermissionForChannel: accessService.HasAnyPermissionForChannel,
		HasPermissionForChannel:    accessService.HasPermissionsForChannel,
		RegisterActor:              registerAction,
		RegisterActorDocumentation: registerActorDocumentation,
		RegisterAPIRoute:           registerRoute,
		RegisterCron:               cronService.AddFunc,
		RegisterCopyDatabaseFunc:   registerDatabaseCopyFunc,
		RegisterEventHandler:       registerEventHandlers,
		RegisterMessageModFunc:     registerChatcommand,
		RegisterRawMessageHandler:  registerRawMessageHandler,
		RegisterTemplateFunction:   tplFuncs.Register,
		SendMessage:                sendMessage,
		ValidateToken:              authService.ValidateTokenFor,

		CreateEvent: func(evt string, eventData *plugins.FieldCollection) error {
			handleMessage(ircHdl.Client(), nil, &evt, eventData)
			return nil
		},

		GetModuleConfigForChannel: func(module, channel string) *plugins.FieldCollection {
			return config.ModuleConfig.GetChannelConfig(module, channel)
		},

		GetTwitchClientForChannel: func(channel string) (*twitch.Client, error) {
			//nolint:wrapcheck // own package, no need to wrap
			return accessService.GetTwitchClientForChannel(channel, access.ClientConfig{
				TwitchClient:       cfg.TwitchClient,
				TwitchClientSecret: cfg.TwitchClientSecret,
			})
		},
	}
}

func sendMessage(m *irc.Message) error {
	err := handleChatcommandModifications(m)
	switch {
	case err == nil:
		// There was no error, the message should be sent normally

	case errors.Is(err, plugins.ErrSkipSendingMessage):
		// One chatcommand handler cancelled sending the message
		// (probably because it was handled otherwise)
		return nil

	default:
		// Something in a chatcommand handler went wrong
		return errors.Wrap(err, "handling chat commands")
	}

	if err = backoff.NewBackoff().WithMaxIterations(ircHandleWaitRetries).Retry(func() error {
		if ircHdl == nil {
			return errors.New("irc handle not available")
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "waiting for IRC connection")
	}

	return ircHdl.SendMessage(m)
}
