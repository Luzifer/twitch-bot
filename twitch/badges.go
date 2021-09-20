package twitch

import (
	"strconv"
	"strings"

	"github.com/go-irc/irc"
)

const (
	BadgeBroadcaster = "broadcaster"
	BadgeFounder     = "founder"
	BadgeModerator   = "moderator"
	BadgeSubscriber  = "subscriber"
	BadgeVIP         = "vip"
)

var KnownBadges = []string{
	BadgeBroadcaster,
	BadgeFounder,
	BadgeModerator,
	BadgeSubscriber,
	BadgeVIP,
}

type BadgeCollection map[string]*int

func ParseBadgeLevels(m *irc.Message) BadgeCollection {
	out := BadgeCollection{}

	if m == nil {
		return out
	}

	badgeString, ok := m.GetTag("badges")
	if !ok || len(badgeString) == 0 {
		return out
	}

	badges := strings.Split(badgeString, ",")
	for _, b := range badges {
		badgeParts := strings.Split(b, "/")
		if len(badgeParts) != 2 { //nolint:gomnd // This is not a magic number but just an expected count
			continue
		}

		level, err := strconv.Atoi(badgeParts[1])
		if err != nil {
			continue
		}

		out.Add(badgeParts[0], level)
	}

	// If there is a founders badge but no subscribers badge
	// add a level-0 subscribers badge to prevent the bot to
	// cause trouble on founders when subscribers are allowed
	// to do something
	if out.Has(BadgeFounder) && !out.Has(BadgeSubscriber) {
		out.Add(BadgeSubscriber, out.Get(BadgeFounder))
	}

	return out
}

func (b BadgeCollection) Add(badge string, level int) {
	b[badge] = &level
}

func (b BadgeCollection) Get(badge string) int {
	l, ok := b[badge]
	if !ok {
		return 0
	}

	return *l
}

func (b BadgeCollection) Has(badge string) bool {
	return b[badge] != nil
}
