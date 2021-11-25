package random

import (
	"crypto/md5" // #nosec G501 // Unly used to convert a string into a numer, no need for cryptographic safety
	"fmt"
	"math"
	"math/rand"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/plugins"
)

func Register(args plugins.RegistrationArguments) error {
	args.RegisterTemplateFunction("seededRandom", plugins.GenericTemplateFunctionGetter(stableRandomFromSeed))
	return nil
}

func stableRandomFromSeed(seed string) (float64, error) {
	seedValue, err := stringToSeed(seed)
	if err != nil {
		return 0, errors.Wrap(err, "generating seed")
	}

	return rand.New(rand.NewSource(seedValue)).Float64(), nil // #nosec G404 // Only used for generating a random number from static string, no need for cryptographic safety
}

func stringToSeed(s string) (int64, error) {
	hash := md5.New() // #nosec G401 // Unly used to convert a string into a numer, no need for cryptographic safety
	if _, err := fmt.Fprint(hash, s); err != nil {
		return 0, errors.Wrap(err, "writing string to hasher")
	}

	var (
		hashSum = hash.Sum(nil)
		sum     int64
	)

	for i := 0; i < len(hashSum); i++ {
		sum += int64(float64(hashSum[len(hashSum)-1-i]%10) * math.Pow(10, float64(i))) //nolint:gomnd // No need to put the 10 of 10**i into a constant named "ten"
	}

	return sum, nil
}
