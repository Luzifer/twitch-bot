package overlays

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Luzifer/go_helpers/v2/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
)

func TestEventDatabaseRoundtrip(t *testing.T) {
	dbc := database.GetTestDatabase(t)
	require.NoError(t, dbc.DB().AutoMigrate(&overlaysEvent{}))

	var (
		channel = "#test"
		evtID   uint64
		tEvent1 = time.Now().UTC()
		tEvent2 = tEvent1.Add(time.Second)
	)

	evts, err := getChannelEvents(dbc, channel)
	assert.NoError(t, err, "getting events on empty db")
	assert.Zero(t, evts, "expect no events on empty db")

	evtID, err = addChannelEvent(dbc, channel, socketMessage{
		IsLive: true,
		Time:   tEvent2,
		Type:   "event 2",
		Fields: fieldcollection.FieldCollectionFromData(map[string]any{"foo": "bar"}),
	})
	assert.Equal(t, uint64(1), evtID)
	assert.NoError(t, err, "adding second event")

	evtID, err = addChannelEvent(dbc, channel, socketMessage{
		IsLive: true,
		Time:   tEvent1,
		Type:   "event 1",
		Fields: fieldcollection.FieldCollectionFromData(map[string]any{"foo": "bar"}),
	})
	assert.Equal(t, uint64(2), evtID)
	assert.NoError(t, err, "adding first event")

	evtID, err = addChannelEvent(dbc, "#otherchannel", socketMessage{
		IsLive: true,
		Time:   tEvent1,
		Type:   "event",
		Fields: fieldcollection.FieldCollectionFromData(map[string]any{"foo": "bar"}),
	})
	assert.Equal(t, uint64(3), evtID)
	assert.NoError(t, err, "adding other channel event")

	evts, err = getChannelEvents(dbc, channel)
	assert.NoError(t, err, "getting events")
	assert.Len(t, evts, 2, "expect 2 events")

	assert.Less(t, evts[0].Time, evts[1].Time, "expect sorting")

	evt, err := getEventByID(dbc, 2)
	assert.NoError(t, err)
	assert.Equal(t, socketMessage{
		EventID: 2,
		IsLive:  false,
		Time:    tEvent1,
		Type:    "event 1",
		Fields:  fieldcollection.FieldCollectionFromData(map[string]any{"foo": "bar"}),
	}, evt)
}
