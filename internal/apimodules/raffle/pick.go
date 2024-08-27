package raffle

import (
	"crypto/rand"
	"encoding/binary"
	mathRand "math/rand"

	"github.com/pkg/errors"
)

type (
	cryptRandSrc struct{}
)

var errNoCandidatesLeft = errors.New("no candidates left")

func pickWinnerFromRaffle(r raffle) (winner raffleEntry, err error) {
	var maxScore float64
	for _, re := range r.Entries {
		if re.WasPicked {
			// We skip previously picked winners and pretend they
			// don't exist
			continue
		}

		maxScore += re.Multiplier
	}

	if maxScore == 0 {
		return winner, errNoCandidatesLeft
	}

	winnerPoint := mathRand.New(cryptRandSrc{}).Float64() * maxScore //#nosec:G404 - RNG is using a secure source

	for i := range r.Entries {
		re := r.Entries[i]

		if re.WasPicked {
			// We skip previously picked winners and pretend they
			// don't exist
			continue
		}

		winnerPoint -= re.Multiplier
		if winnerPoint < 0 {
			winner = re
			break
		}
	}

	return winner, nil
}

func (cryptRandSrc) Int63() int64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return -1
	}
	// mask off sign bit to ensure positive number
	return int64(binary.LittleEndian.Uint64(b[:]) & (1<<63 - 1)) //#nosec:G115 - Masking ensures conversion is fine
}

// We're using a non-seedable source
func (cryptRandSrc) Seed(int64) {}
