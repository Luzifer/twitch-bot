package overlays

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

func TestEventDatabaseRoundtrip(t *testing.T) {
	dbc := database.GetTestDatabase(t)
	require.NoError(t, dbc.DB().AutoMigrate(&overlaysEvent{}))

	var (
		channel = "#test"
		tEvent1 = time.Now()
		tEvent2 = tEvent1.Add(time.Second)
	)

	evts, err := GetChannelEvents(dbc, channel)
	assert.NoError(t, err, "getting events on empty db")
	assert.Zero(t, evts, "expect no events on empty db")

	assert.NoError(t, AddChannelEvent(dbc, channel, SocketMessage{
		IsLive: true,
		Time:   tEvent2,
		Type:   "event 2",
		Fields: plugins.FieldCollectionFromData(map[string]any{"foo": "bar"}),
	}), "adding second event")

	assert.NoError(t, AddChannelEvent(dbc, channel, SocketMessage{
		IsLive: true,
		Time:   tEvent1,
		Type:   "event 1",
		Fields: plugins.FieldCollectionFromData(map[string]any{"foo": "bar"}),
	}), "adding first event")

	assert.NoError(t, AddChannelEvent(dbc, "#otherchannel", SocketMessage{
		IsLive: true,
		Time:   tEvent1,
		Type:   "event",
		Fields: plugins.FieldCollectionFromData(map[string]any{"foo": "bar"}),
	}), "adding other channel event")

	evts, err = GetChannelEvents(dbc, channel)
	assert.NoError(t, err, "getting events")
	assert.Len(t, evts, 2, "expect 2 events")

	assert.Less(t, evts[0].Time, evts[1].Time, "expect sorting")
}
