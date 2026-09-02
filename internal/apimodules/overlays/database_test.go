package overlays

import (
	"testing"
	"time"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
)

func TestEventDatabaseRetentionDelete(t *testing.T) {
	dbc := database.GetTestDatabase(t)
	require.NoError(t, dbc.DB().AutoMigrate(&overlaysEvent{}))

	var (
		channel      = "#retention-test"
		cutoff       = time.Date(2026, time.September, 2, 12, 0, 0, 0, time.UTC)
		eventFields  = fieldcollection.FromData(map[string]any{"foo": "bar"})
		otherChannel = "#retention-otherchannel"
	)

	for _, evt := range []struct {
		channel   string
		eventType string
		time      time.Time
	}{
		{channel, "old", cutoff.Add(-time.Second)},
		{channel, "at cutoff", cutoff},
		{channel, "new", cutoff.Add(time.Second)},
		{otherChannel, "other channel old", cutoff.Add(-time.Second)},
	} {
		_, err := addChannelEvent(dbc, evt.channel, socketMessage{
			Time:   evt.time,
			Type:   evt.eventType,
			Fields: eventFields,
		})
		require.NoError(t, err, "adding event %q", evt.eventType)
	}

	require.NoError(t, deleteEventsOlderThan(dbc, channel, cutoff))

	evts, err := getChannelEvents(dbc, channel)
	require.NoError(t, err, "getting retained channel events")
	require.Len(t, evts, 2)
	assert.Equal(t, "at cutoff", evts[0].Type)
	assert.Equal(t, "new", evts[1].Type)

	evts, err = getChannelEvents(dbc, otherChannel)
	require.NoError(t, err, "getting other channel events")
	require.Len(t, evts, 1)
	assert.Equal(t, "other channel old", evts[0].Type)
}

func TestEventDatabaseRoundtrip(t *testing.T) {
	dbc := database.GetTestDatabase(t)
	require.NoError(t, dbc.DB().AutoMigrate(&overlaysEvent{}))

	var (
		channel      = "#roundtrip-test"
		event1ID     uint64
		evtID        uint64
		tEvent1      = time.Now().UTC()
		tEvent2      = tEvent1.Add(time.Second)
		otherChannel = "#roundtrip-otherchannel"
	)

	evts, err := getChannelEvents(dbc, channel)
	require.NoError(t, err, "getting events on empty db")
	assert.Zero(t, evts, "expect no events on empty db")

	evtID, err = addChannelEvent(dbc, channel, socketMessage{
		IsLive: true,
		Time:   tEvent2,
		Type:   "event 2",
		Fields: fieldcollection.FromData(map[string]any{"foo": "bar"}),
	})
	require.NoError(t, err, "adding second event")
	assert.NotZero(t, evtID)

	event1ID, err = addChannelEvent(dbc, channel, socketMessage{
		IsLive: true,
		Time:   tEvent1,
		Type:   "event 1",
		Fields: fieldcollection.FromData(map[string]any{"foo": "bar"}),
	})
	require.NoError(t, err, "adding first event")
	assert.NotZero(t, event1ID)

	evtID, err = addChannelEvent(dbc, otherChannel, socketMessage{
		IsLive: true,
		Time:   tEvent1,
		Type:   "event",
		Fields: fieldcollection.FromData(map[string]any{"foo": "bar"}),
	})
	require.NoError(t, err, "adding other channel event")
	assert.NotZero(t, evtID)

	evts, err = getChannelEvents(dbc, channel)
	require.NoError(t, err, "getting events")
	assert.Len(t, evts, 2, "expect 2 events")

	assert.Less(t, evts[0].Time, evts[1].Time, "expect sorting")

	evt, err := getEventByID(dbc, event1ID)
	require.NoError(t, err)
	assert.Equal(t, socketMessage{
		EventID: event1ID,
		IsLive:  false,
		Time:    tEvent1,
		Type:    "event 1",
		Fields:  fieldcollection.FromData(map[string]any{"foo": "bar"}),
	}, evt)
}
