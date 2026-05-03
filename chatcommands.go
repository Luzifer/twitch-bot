package main

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var (
	availableChatcommands     = make(map[string]plugins.MsgModificationFunc)
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
			return fmt.Errorf("modifying message: %w", err)
		}
	}

	return nil
}
