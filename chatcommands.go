package main

import (
	"strings"
	"sync"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v2/plugins"
)

var (
	availableChatcommands     = map[string]plugins.MsgModificationFunc{}
	availableChatcommandsLock = new(sync.RWMutex)
)

func registerChatcommand(linePrefix string, modFn plugins.MsgModificationFunc) {
	availableChatcommandsLock.Lock()
	defer availableChatcommandsLock.Unlock()

	if _, ok := availableChatcommands[linePrefix]; ok {
		log.WithField("linePrefix", linePrefix).Fatal("Duplicate registration of chatcommand")
	}

	availableChatcommands[linePrefix] = modFn
}

func handleChatcommandModifications(m *irc.Message) error {
	availableChatcommandsLock.RLock()
	defer availableChatcommandsLock.RUnlock()

	msg := m.Trailing()

	for prefix, modFn := range availableChatcommands {
		if !strings.HasPrefix(msg, prefix) {
			continue
		}

		if err := modFn(m); err != nil {
			return errors.Wrap(err, "modifying message")
		}
	}

	return nil
}
