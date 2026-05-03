package timer

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

// AddPermit adds a new permit timer
func (s Service) AddPermit(channel, username string) error {
	return s.SetTimer(s.getPermitTimerKey(channel, username), time.Now().Add(s.permitTimeout))
}

// HasPermit checks whether a valid permit is present
func (s Service) HasPermit(channel, username string) (bool, error) {
	return s.HasTimer(s.getPermitTimerKey(channel, username))
}

func (Service) getPermitTimerKey(channel, username string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(fmt.Appendf(nil,
		"%d:%s:%s",
		plugins.TimerTypePermit, channel, strings.ToLower(strings.TrimLeft(username, "@")),
	)))
}
