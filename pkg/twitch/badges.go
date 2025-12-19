package twitch

import (
	"strconv"
	"strings"

	"gopkg.in/irc.v4"
)

// Collection of known badges
const (
	BadgeBroadcaster   = "broadcaster"
	BadgeFounder       = "founder"
	BadgeLeadModerator = "lead_moderator"
	BadgeModerator     = "moderator"
	BadgeSubscriber    = "subscriber"
	BadgeVIP           = "vip"
)

// KnownBadges contains a list of all known badges
var KnownBadges = []string{
	BadgeBroadcaster,
	BadgeFounder,
	BadgeLeadModerator,
	BadgeModerator,
	BadgeSubscriber,
	BadgeVIP,
}

// BadgeCollection represents a collection of badges the user has set
type BadgeCollection map[string]*int

// ParseBadgeLevels takes the badges from the irc.Message and returns
// a BadgeCollection containing all badges the user has set
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
		if len(badgeParts) != 2 { //nolint:mnd // This is not a magic number but just an expected count
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

	// Twitch introduced Lead-Moderators which take the same
	// badge slot as normal moderators. For simplicity sake
	// we grant every lead-moderator also moderator badge so
	// when a moderator can do stuff, the lead-mod can do the
	// same.
	if out.Has(BadgeLeadModerator) && !out.Has(BadgeModerator) {
		out.Add(BadgeModerator, 1)
	}

	return out
}

// Add sets the given badge to the given level
func (b BadgeCollection) Add(badge string, level int) {
	b[badge] = &level
}

// Get returns the level of the given badge. If the badge is not set
// its level will be 0.
func (b BadgeCollection) Get(badge string) int {
	l, ok := b[badge]
	if !ok {
		return 0
	}

	return *l
}

// Has checks whether the collection contains the given badge at any
// level
func (b BadgeCollection) Has(badge string) bool {
	return b[badge] != nil
}
