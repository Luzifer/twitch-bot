package timer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
)

func TestTimerRoundtrip(t *testing.T) {
	dbc := database.GetTestDatabase(t)
	ts, err := New(dbc, nil)
	require.NoError(t, err, "creating timer service")

	id := "78c0176a-938e-497b-bed4-83d5bdec6caf"

	has, err := ts.HasTimer(id)
	require.NoError(t, err, "checking for non-existent timer")
	assert.False(t, has, "checking existence of non-existing timer")

	err = ts.SetTimer(id, time.Now().Add(500*time.Millisecond))
	require.NoError(t, err, "setting timer")

	has, err = ts.HasTimer(id)
	require.NoError(t, err, "checking for existent timer")
	assert.True(t, has, "checking existence of existing timer")

	err = ts.SetTimer(id, time.Now().Add(-time.Millisecond))
	require.NoError(t, err, "updating timer")

	has, err = ts.HasTimer(id)
	require.NoError(t, err, "checking for expired timer")
	assert.False(t, has, "checking existence of expired timer")
}
