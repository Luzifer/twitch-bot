package main

import (
	"fmt"
	"net/http"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/go_helpers/v2/backoff"
	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/announce"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/ban"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/counter"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/delay"
	deleteactor "github.com/Luzifer/twitch-bot/v3/internal/actors/delete"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/eventmod"
	"github.com/Luzifer/twitch-bot/v3/internal/actors/filesay"
	logActor "github.com/Luzifer/twitch-bot/v3/internal/actors/log"
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
	"github.com/Luzifer/twitch-bot/v3/internal/service/access"
	"github.com/Luzifer/twitch-bot/v3/internal/template/api"
	"github.com/Luzifer/twitch-bot/v3/internal/template/numeric"
	"github.com/Luzifer/twitch-bot/v3/internal/template/random"
	"github.com/Luzifer/twitch-bot/v3/internal/template/slice"
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
		counter.Register,
		delay.Register,
		deleteactor.Register,
		eventmod.Register,
		filesay.Register,
		logActor.Register,
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

		// API-only modules
		customevent.Register,
		msgformat.Register,
		overlays.Register,
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
		hdl = botEditorAuthMiddleware(hdl)
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
		CreateEvent: func(evt string, eventData *plugins.FieldCollection) error {
			handleMessage(ircHdl.Client(), nil, &evt, eventData)
			return nil
		},
		FormatMessage:              formatMessage,
		GetDatabaseConnector:       func() database.Connector { return db },
		GetLogger:                  func(moduleName string) *log.Entry { return log.WithField("module", moduleName) },
		GetTwitchClient:            func() *twitch.Client { return twitchClient },
		HasAnyPermissionForChannel: accessService.HasAnyPermissionForChannel,
		HasPermissionForChannel:    accessService.HasPermissionsForChannel,
		RegisterActor:              registerAction,
		RegisterActorDocumentation: registerActorDocumentation,
		RegisterAPIRoute:           registerRoute,
		RegisterCron:               cronService.AddFunc,
		RegisterEventHandler:       registerEventHandlers,
		RegisterMessageModFunc:     registerChatcommand,
		RegisterRawMessageHandler:  registerRawMessageHandler,
		RegisterTemplateFunction:   tplFuncs.Register,
		SendMessage:                sendMessage,
		ValidateToken:              validateAuthToken,

		GetTwitchClientForChannel: func(channel string) (*twitch.Client, error) {
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
