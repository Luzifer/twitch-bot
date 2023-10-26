package userstate

import (
	"strings"
	"sync"

	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
)

type (
	twitchUserState struct {
		Badges      twitch.BadgeCollection
		Color       string
		DisplayName string
		EmoteSets   []string
	}

	twitchUserStateStore struct {
		states map[string]*twitchUserState
		lock   sync.RWMutex
	}
)

func newTwitchUserStateStore() *twitchUserStateStore {
	return &twitchUserStateStore{
		states: make(map[string]*twitchUserState),
	}
}

func parseTwitchUserState(m *irc.Message) (*twitchUserState, error) {
	var (
		color, _       = m.Tags["color"]
		displayName, _ = m.Tags["display-name"]
		emoteSets      []string
		rawSets, _     = m.Tags["emote-sets"]
	)

	if rawSets != "" {
		emoteSets = strings.Split(rawSets, ",")
	}

	return &twitchUserState{
		Badges:      twitch.ParseBadgeLevels(m),
		Color:       color,
		DisplayName: displayName,
		EmoteSets:   emoteSets,
	}, nil
}

func (t *twitchUserStateStore) Get(channel string) *twitchUserState {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.states[channel]
}

func (t *twitchUserStateStore) Set(channel string, state *twitchUserState) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.states[channel] = state
}
