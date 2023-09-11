package twitch

import (
	"strconv"
	"strings"

	"gopkg.in/irc.v4"
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

	badgeString, ok := m.Tags["badges"]
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

	// In order to simplify queries and permissions add an
	// implicit moderator badge to broadcasters as they have
	// the same (and more) permissions than a moderator. So
	// when allowing actions for moderators now broadcasters
	// ill also be included.
	if out.Has(BadgeBroadcaster) && !out.Has(BadgeModerator) {
		out.Add(BadgeModerator, 1)
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
