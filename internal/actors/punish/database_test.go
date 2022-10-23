package punish

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Luzifer/twitch-bot/v2/pkg/database"
)

func TestPunishmentRoundtrip(t *testing.T) {
	dbc := database.GetTestDatabase(t)
	require.NoError(t, dbc.DB().AutoMigrate(&punishLevel{}))

	var (
		channel = "#test"
		user    = "test"
		uuid    = "1befb33d-be89-4724-8ae1-0d465eb58007"
	)

	pl, err := getPunishment(dbc, channel, user, uuid)
	assert.NoError(t, err, "query non-existent punishment")
	assert.Equal(t, -1, pl.LastLevel, "check default level")
	assert.Zero(t, pl.Executed, "check default time")
	assert.Zero(t, pl.Cooldown, "check default cooldown")

	err = setPunishment(dbc, channel, user, uuid, &levelConfig{
		Cooldown:  500 * time.Millisecond,
		Executed:  time.Now().UTC(),
		LastLevel: 1,
	})
	assert.NoError(t, err, "setting punishment")

	pl, err = getPunishment(dbc, channel, user, uuid)
	assert.NoError(t, err, "query existent punishment")
	assert.Equal(t, 1, pl.LastLevel, "check level without cooldown")

	time.Sleep(500 * time.Millisecond) // Wait for one cooldown to happen

	pl, err = getPunishment(dbc, channel, user, uuid)
	assert.NoError(t, err, "query existent punishment")
	assert.Equal(t, 0, pl.LastLevel, "check level after one cooldown")
	assert.NotZero(t, pl.Executed, "check non-zero-time after one cooldown")
	assert.Equal(t, 500*time.Millisecond, pl.Cooldown, "check non-zero-cooldown after one cooldown")

	time.Sleep(500 * time.Millisecond) // Wait for one cooldown to happen

	pl, err = getPunishment(dbc, channel, user, uuid)
	assert.NoError(t, err, "query existent punishment")
	assert.Equal(t, -1, pl.LastLevel, "check level after two cooldown")
	assert.Zero(t, pl.Executed, "check zero-time after two cooldown")
	assert.Zero(t, pl.Cooldown, "check zero-cooldown after two cooldown")
}
