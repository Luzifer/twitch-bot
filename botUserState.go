package main

import (
	"strconv"
	"strings"
	"sync"

	"github.com/Luzifer/twitch-bot/twitch"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

type (
	twitchUserState struct {
		Badges      twitch.BadgeCollection
		Color       string
		DisplayName string
		EmoteSets   []int64
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
		color, _       = m.GetTag("color")
		displayName, _ = m.GetTag("display-name")
		emoteSets      []int64
		rawSets, _     = m.GetTag("emote-sets")
	)

	if rawSets != "" {
		for _, sid := range strings.Split(rawSets, ",") {
			id, err := strconv.ParseInt(sid, 10, 64)
			if err != nil {
				return nil, errors.Wrap(err, "parsing emote-set id")
			}
			emoteSets = append(emoteSets, id)
		}
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
