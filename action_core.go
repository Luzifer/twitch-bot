package main

import (
	"fmt"

	"github.com/Luzifer/twitch-bot/internal/actors/ban"
	"github.com/Luzifer/twitch-bot/internal/actors/delay"
	deleteactor "github.com/Luzifer/twitch-bot/internal/actors/delete"
	"github.com/Luzifer/twitch-bot/internal/actors/modchannel"
	"github.com/Luzifer/twitch-bot/internal/actors/raw"
	"github.com/Luzifer/twitch-bot/internal/actors/respond"
	"github.com/Luzifer/twitch-bot/internal/actors/timeout"
	"github.com/Luzifer/twitch-bot/internal/actors/whisper"
	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/Luzifer/twitch-bot/twitch"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var coreActorRegistations = []plugins.RegisterFunc{
	ban.Register,
	delay.Register,
	deleteactor.Register,
	modchannel.Register,
	raw.Register,
	respond.Register,
	timeout.Register,
	whisper.Register,
}

func initCorePlugins() error {
	args := getRegistrationArguments()
	for _, rf := range coreActorRegistations {
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

	if route.IsPrefix {
		r.PathPrefix(route.Path).
			HandlerFunc(route.HandlerFunc).
			Methods(route.Method)
	} else {
		r.HandleFunc(route.Path, route.HandlerFunc).
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
