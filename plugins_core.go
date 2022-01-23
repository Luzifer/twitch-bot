package main

import (
	"fmt"
	"net/http"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/go_helpers/v2/backoff"
	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/internal/actors/ban"
	"github.com/Luzifer/twitch-bot/internal/actors/delay"
	deleteactor "github.com/Luzifer/twitch-bot/internal/actors/delete"
	"github.com/Luzifer/twitch-bot/internal/actors/modchannel"
	"github.com/Luzifer/twitch-bot/internal/actors/nuke"
	"github.com/Luzifer/twitch-bot/internal/actors/punish"
	"github.com/Luzifer/twitch-bot/internal/actors/quotedb"
	"github.com/Luzifer/twitch-bot/internal/actors/raw"
	"github.com/Luzifer/twitch-bot/internal/actors/respond"
	"github.com/Luzifer/twitch-bot/internal/actors/timeout"
	"github.com/Luzifer/twitch-bot/internal/actors/whisper"
	"github.com/Luzifer/twitch-bot/internal/apimodules/msgformat"
	"github.com/Luzifer/twitch-bot/internal/apimodules/overlays"
	"github.com/Luzifer/twitch-bot/internal/template/numeric"
	"github.com/Luzifer/twitch-bot/internal/template/random"
	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/Luzifer/twitch-bot/twitch"
)

const ircHandleWaitRetries = 10

var (
	corePluginRegistrations = []plugins.RegisterFunc{
		// Actors
		ban.Register,
		delay.Register,
		deleteactor.Register,
		modchannel.Register,
		nuke.Register,
		punish.Register,
		quotedb.Register,
		raw.Register,
		respond.Register,
		timeout.Register,
		whisper.Register,

		// Template functions
		numeric.Register,
		random.Register,

		// API-only modules
		msgformat.Register,
		overlays.Register,
	}
	knownModules []string
)

func initCorePlugins() error {
	args := getRegistrationArguments()
	for _, rf := range corePluginRegistrations {
		if err := rf(args); err != nil {
			return errors.Wrap(err, "registering core plugin")
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
		FormatMessage:              formatMessage,
		GetLogger:                  func(moduleName string) *log.Entry { return log.WithField("module", moduleName) },
		GetStorageManager:          func() plugins.StorageManager { return store },
		GetTwitchClient:            func() *twitch.Client { return twitchClient },
		RegisterActor:              registerAction,
		RegisterActorDocumentation: registerActorDocumentation,
		RegisterAPIRoute:           registerRoute,
		RegisterCron:               cronService.AddFunc,
		RegisterRawMessageHandler:  registerRawMessageHandler,
		RegisterTemplateFunction:   tplFuncs.Register,
		SendMessage:                sendMessage,
	}
}

func sendMessage(m *irc.Message) error {
	if err := backoff.NewBackoff().WithMaxIterations(ircHandleWaitRetries).Retry(func() error {
		if ircHdl == nil {
			return errors.New("irc handle not available")
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "waiting for IRC connection")
	}

	return ircHdl.SendMessage(m)
}
