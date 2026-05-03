package counter

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
)

func TestCounterStoreLoop(t *testing.T) {
	dbc := database.GetTestDatabase(t)
	require.NoError(t, dbc.DB().AutoMigrate(&counter{}))

	counterName := "mytestcounter"

	v, err := getCounterValue(dbc, counterName)
	require.NoError(t, err, "reading non-existent counter")
	assert.Equal(t, int64(0), v, "expecting 0 counter value on non-existent counter")

	err = updateCounter(dbc, counterName, 5, true, time.Now())
	require.NoError(t, err, "inserting counter")

	var rawCounter counter
	require.NoError(t, dbc.DB().First(&rawCounter, "name = ?", counterName).Error)
	assert.Equal(t, rawCounter.FirstSeen, rawCounter.LastModified)

	err = updateCounter(dbc, counterName, 1, false, time.Now())
	require.NoError(t, err, "updating counter")

	require.NoError(t, dbc.DB().First(&rawCounter, "name = ?", counterName).Error)
	assert.NotEqual(t, rawCounter.FirstSeen, rawCounter.LastModified)

	v, err = getCounterValue(dbc, counterName)
	require.NoError(t, err, "reading existent counter")
	assert.Equal(t, int64(6), v, "expecting counter value on existing counter")
}

func TestCounterTopListAndRank(t *testing.T) {
	dbc := database.GetTestDatabase(t)
	require.NoError(t, dbc.DB().AutoMigrate(&counter{}))

	testTime := time.Now().UTC()

	counterTemplate := `#example:test:%v`
	for i := range 6 {
		require.NoError(
			t,
			updateCounter(dbc, fmt.Sprintf(counterTemplate, i), int64(i), true, testTime),
			"inserting counter %d", i,
		)
	}

	cc, err := getCounterTopList(dbc, fmt.Sprintf(counterTemplate, ""), 3)
	require.NoError(t, err)
	assert.Len(t, cc, 3)

	assert.Equal(t, []counter{
		{Name: "#example:test:5", Value: 5, FirstSeen: testTime, LastModified: testTime},
		{Name: "#example:test:4", Value: 4, FirstSeen: testTime, LastModified: testTime},
		{Name: "#example:test:3", Value: 3, FirstSeen: testTime, LastModified: testTime},
	}, cc)

	cc, err = getCounterTopList(dbc, fmt.Sprintf(counterTemplate, ""), 3, "name DESC")
	require.NoError(t, err)
	assert.Len(t, cc, 3)

	assert.Equal(t, []counter{
		{Name: "#example:test:5", Value: 5, FirstSeen: testTime, LastModified: testTime},
		{Name: "#example:test:4", Value: 4, FirstSeen: testTime, LastModified: testTime},
		{Name: "#example:test:3", Value: 3, FirstSeen: testTime, LastModified: testTime},
	}, cc)

	cc, err = getCounterTopList(dbc, fmt.Sprintf(counterTemplate, ""), 3, "name")
	require.NoError(t, err)
	assert.Len(t, cc, 3)

	assert.Equal(t, []counter{
		{Name: "#example:test:0", Value: 0, FirstSeen: testTime, LastModified: testTime},
		{Name: "#example:test:1", Value: 1, FirstSeen: testTime, LastModified: testTime},
		{Name: "#example:test:2", Value: 2, FirstSeen: testTime, LastModified: testTime},
	}, cc)

	_, err = getCounterTopList(dbc, fmt.Sprintf(counterTemplate, ""), 3, "foobar")
	require.Error(t, err)

	_, err = getCounterTopList(dbc, fmt.Sprintf(counterTemplate, ""), 3, "name foo")
	require.Error(t, err)

	_, err = getCounterTopList(dbc, fmt.Sprintf(counterTemplate, ""), 3, "name ASC; DROP TABLE counters;")
	require.Error(t, err)

	rank, count, err := getCounterRank(dbc,
		fmt.Sprintf(counterTemplate, ""),
		fmt.Sprintf(counterTemplate, 4))
	require.NoError(t, err)
	assert.Equal(t, int64(6), count)
	assert.Equal(t, int64(2), rank)
}
