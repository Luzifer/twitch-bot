package raffle

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testGenerateRaffe() raffle {
	r := raffle{
		MultiFollower:   1.1,
		MultiSubscriber: 1.2,
		MultiVIP:        1.5,

		Entries: make([]raffleEntry, 0, 837),
	}

	// Now lets generate 132 non-followers taking part
	for i := 0; i < 132; i++ {
		r.Entries = append(r.Entries, raffleEntry{ID: uint64(i), Multiplier: 1})
	}

	// Now lets generate 500 followers taking part
	for i := 0; i < 500; i++ {
		r.Entries = append(r.Entries, raffleEntry{ID: 10000 + uint64(i), Multiplier: r.MultiFollower})
	}

	// Now lets generate 200 subscribers taking part
	for i := 0; i < 200; i++ {
		r.Entries = append(r.Entries, raffleEntry{ID: 20000 + uint64(i), Multiplier: r.MultiSubscriber})
	}

	// Now lets generate 5 VIPs taking part
	for i := 0; i < 5; i++ {
		r.Entries = append(r.Entries, raffleEntry{ID: 30000 + uint64(i), Multiplier: r.MultiVIP})
	}

	// They didn't join in order so lets shuffle them
	rand.Shuffle(len(r.Entries), func(i, j int) { r.Entries[i], r.Entries[j] = r.Entries[j], r.Entries[i] })

	return r
}

func BenchmarkPickWinnerFromRaffle(b *testing.B) {
	tData := testGenerateRaffe()
	var err error

	b.Run("pick", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err = pickWinnerFromRaffle(tData)
			require.NoError(b, err)
		}
	})
}

func TestPickWinnerFromRaffle(t *testing.T) {
	var (
		winners []uint64
		tData   = testGenerateRaffe()
	)

	for i := 0; i < 5; i++ {
		w, err := pickWinnerFromRaffle(tData)
		require.NoError(t, err, "picking winner")
		winners = append(winners, w.ID)
	}

	t.Logf("winners: %v", winners)
}

func TestPickWinnerFromRaffleSpecial(t *testing.T) {
	r := raffle{}
	_, err := pickWinnerFromRaffle(r)
	assert.ErrorIs(t, errNoCandidatesLeft, err, "picking from 0 paricipants")

	r.Entries = append(r.Entries, raffleEntry{ID: 1, Multiplier: 1.0})
	winner, err := pickWinnerFromRaffle(r)
	assert.NoError(t, err, "picking from set of 1")
	assert.Equal(t, uint64(1), winner.ID, "expect the right winner")

	r.Entries[0].WasPicked = true
	_, err = pickWinnerFromRaffle(r)
	assert.ErrorIs(t, errNoCandidatesLeft, err, "picking from 1 paricipant, which already won")
}
