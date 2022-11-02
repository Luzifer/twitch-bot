package counter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
)

func TestCounterStoreLoop(t *testing.T) {
	dbc := database.GetTestDatabase(t)
	dbc.DB().AutoMigrate(&counter{})

	counterName := "mytestcounter"

	v, err := GetCounterValue(dbc, counterName)
	assert.NoError(t, err, "reading non-existent counter")
	assert.Equal(t, int64(0), v, "expecting 0 counter value on non-existent counter")

	err = UpdateCounter(dbc, counterName, 5, true)
	assert.NoError(t, err, "inserting counter")

	err = UpdateCounter(dbc, counterName, 1, false)
	assert.NoError(t, err, "updating counter")

	v, err = GetCounterValue(dbc, counterName)
	assert.NoError(t, err, "reading existent counter")
	assert.Equal(t, int64(6), v, "expecting counter value on existing counter")
}
