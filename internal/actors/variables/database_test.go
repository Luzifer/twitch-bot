package variables

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Luzifer/twitch-bot/v2/pkg/database"
)

func TestVariableRoundtrip(t *testing.T) {
	dbc := database.GetTestDatabase(t)
	require.NoError(t, dbc.DB().AutoMigrate(&variable{}), "applying migration")

	var (
		name      = "myvar"
		testValue = "ee5e4be5-f292-48aa-a177-cb9fd6f4e171"
	)

	v, err := GetVariable(dbc, name)
	assert.NoError(t, err, "getting unset variable")
	assert.Zero(t, v, "checking zero state on unset variable")

	assert.NoError(t, SetVariable(dbc, name, testValue), "setting variable")

	v, err = GetVariable(dbc, name)
	assert.NoError(t, err, "getting set variable")
	assert.NotZero(t, v, "checking non-zero state on set variable")

	assert.NoError(t, RemoveVariable(dbc, name), "removing variable")

	v, err = GetVariable(dbc, name)
	assert.NoError(t, err, "getting removed variable")
	assert.Zero(t, v, "checking zero state on removed variable")
}
