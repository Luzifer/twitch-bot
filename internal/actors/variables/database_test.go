package variables

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
)

func TestVariableRoundtrip(t *testing.T) {
	dbc := database.GetTestDatabase(t)
	require.NoError(t, dbc.DB().AutoMigrate(&variable{}), "applying migration")

	var (
		name      = "myvar"
		testValue = "ee5e4be5-f292-48aa-a177-cb9fd6f4e171"
	)

	v, err := getVariable(dbc, name)
	require.NoError(t, err, "getting unset variable")
	assert.Empty(t, v, "checking zero state on unset variable")

	require.NoError(t, setVariable(dbc, name, testValue), "setting variable")

	v, err = getVariable(dbc, name)
	require.NoError(t, err, "getting set variable")
	assert.NotEmpty(t, v, "checking non-zero state on set variable")

	require.NoError(t, removeVariable(dbc, name), "removing variable")

	v, err = getVariable(dbc, name)
	require.NoError(t, err, "getting removed variable")
	assert.Empty(t, v, "checking zero state on removed variable")
}
