package main

import (
	"github.com/Luzifer/twitch-bot/internal/actors/ban"
	"github.com/Luzifer/twitch-bot/internal/actors/delay"
	deleteactor "github.com/Luzifer/twitch-bot/internal/actors/delete"
	"github.com/Luzifer/twitch-bot/internal/actors/raw"
	"github.com/Luzifer/twitch-bot/internal/actors/respond"
	"github.com/Luzifer/twitch-bot/internal/actors/timeout"
	"github.com/Luzifer/twitch-bot/internal/actors/whisper"
	"github.com/Luzifer/twitch-bot/plugins"
	log "github.com/sirupsen/logrus"
)

var coreActorRegistations = []plugins.RegisterFunc{
	ban.Register,
	delay.Register,
	deleteactor.Register,
	raw.Register,
	respond.Register,
	timeout.Register,
	whisper.Register,
}

func init() {
	args := getRegistrationArguments()
	for _, rf := range coreActorRegistations {
		if err := rf(args); err != nil {
			log.WithError(err).Fatal("Unable to register core actor")
		}
	}
}

func getRegistrationArguments() plugins.RegistrationArguments {
	return plugins.RegistrationArguments{
		FormatMessage: formatMessage,
		RegisterActor: registerAction,
	}
}
