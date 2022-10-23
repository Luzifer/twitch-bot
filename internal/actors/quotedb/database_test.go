package quotedb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Luzifer/twitch-bot/v2/pkg/database"
)

func TestQuotesRoundtrip(t *testing.T) {
	dbc := database.GetTestDatabase(t)
	require.NoError(t, dbc.DB().AutoMigrate(&quote{}))

	var (
		channel = "#test"
		quotes  = []string{
			"I'm a quote",
			"I might have been said...",
			"Testing rocks",
			"Lets add some more",
			"...or not",
		}
	)

	cq, err := GetChannelQuotes(dbc, channel)
	assert.NoError(t, err, "querying empty database")
	assert.Zero(t, cq, "expecting no quotes")

	for i, q := range quotes {
		assert.NoError(t, AddQuote(dbc, channel, q), "adding quote %d", i)
	}

	cq, err = GetChannelQuotes(dbc, channel)
	assert.NoError(t, err, "querying database")
	assert.Equal(t, quotes, cq, "checkin order and presence of quotes")

	assert.NoError(t, DelQuote(dbc, channel, 1), "removing one quote")
	assert.NoError(t, DelQuote(dbc, channel, 1), "removing one quote")

	cq, err = GetChannelQuotes(dbc, channel)
	assert.NoError(t, err, "querying database")
	assert.Len(t, cq, len(quotes)-2, "expecting quotes in db")

	assert.NoError(t, SetQuotes(dbc, channel, quotes), "replacing quotes")

	cq, err = GetChannelQuotes(dbc, channel)
	assert.NoError(t, err, "querying database")
	assert.Equal(t, quotes, cq, "checkin order and presence of quotes")

	idx, q, err := GetQuote(dbc, channel, 0)
	assert.NoError(t, err, "getting random quote")
	assert.NotZero(t, idx, "index must not be zero")
	assert.NotZero(t, q, "quote must not be zero")

	idx, q, err = GetQuote(dbc, channel, 3)
	assert.NoError(t, err, "getting specific quote")
	assert.Equal(t, 3, idx, "index must be 3")
	assert.Equal(t, quotes[2], q, "quote must not the third")
}
