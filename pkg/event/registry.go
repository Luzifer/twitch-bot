package event

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

type (
	// DocumentedEvent is a superset of an Event and can return required
	// parameters for documentation-generators
	DocumentedEvent interface {
		Event
		// Description returns a human-readable description of the event
		Description() string
	}
)

var knownEvents = make(map[string]DocumentedEvent)

// KnownEvents returns all registered documented events sorted by event-name
func KnownEvents() (out []DocumentedEvent) {
	keys := slices.Collect(maps.Keys(knownEvents))
	slices.Sort(keys)

	for _, k := range keys {
		out = append(out, knownEvents[k])
	}

	return out
}

// RegisterDocumentedEvent adds events to the documentation registry
func RegisterDocumentedEvent(evts ...DocumentedEvent) error {
	var errs []error
	for _, evt := range evts {
		name := evt.Event()
		if name == nil {
			errs = append(errs, fmt.Errorf("event must not have nil event-name"))
			continue
		}

		if _, ok := knownEvents[*name]; ok {
			errs = append(errs, fmt.Errorf("event %q was registered twice", *name))
			continue
		}

		knownEvents[*name] = evt
	}

	return errors.Join(errs...)
}

func init() {
	if err := RegisterDocumentedEvent(
		AdBreakBegin{},
		Announcement{},
		Ban{},
		Bits{},
		CategoryUpdate{},
		ChannelPointRedeem{},
		ClearChat{},
		ClearMessage{},
		Follow{},
		GiftPaidUpgrade{},
		HypetrainBegin{},
		HypetrainEnd{},
		HypetrainProgress{},
		Join{},
		ModeratorAdd{},
		ModeratorRemove{},
		OutboundRaid{},
		Part{},
		Permit{},
		PollBegin{},
		PollEnd{},
		PollProgress{},
		PrimePaidUpgrade{},
		Raid{},
		Resub{},
		ShoutoutCreated{},
		ShoutoutReceived{},
		StreamOffline{},
		StreamOnline{},
		Sub{},
		SubGift{},
		SubMysteryGift{},
		SuspiciousUserMessage{},
		SuspiciousUserUpdate{},
		Timeout{},
		TitleUpdate{},
		VIPAdd{},
		VIPRemove{},
		WatchStreak{},
		Whisper{},
	); err != nil {
		panic(err)
	}
}
