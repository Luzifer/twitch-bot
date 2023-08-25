package random

import (
	"crypto/md5" // #nosec G501 // Unly used to convert a string into a numer, no need for cryptographic safety
	"fmt"
	"math"
	"math/rand"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

func Register(args plugins.RegistrationArguments) error {
	args.RegisterTemplateFunction("randomString", plugins.GenericTemplateFunctionGetter(randomString), plugins.TemplateFuncDocumentation{
		Description: "Randomly picks a string from a list of strings",
		Syntax:      "randomString <string> [...string]",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ randomString "a" "b" "c" "d" }}`,
			FakedOutput: "a",
		},
	})

	args.RegisterTemplateFunction("seededRandom", plugins.GenericTemplateFunctionGetter(stableRandomFromSeed), plugins.TemplateFuncDocumentation{
		Description: "Returns a float value stable for the given seed",
		Syntax:      "seededRandom <string-seed>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template: `Your int this hour: {{ printf "%.0f" (mulf (seededRandom (list "int" .username (now | date "2006-01-02 15") | join ":")) 100) }}%`,
		},
	})

	return nil
}

func randomString(lst ...string) (string, error) {
	if len(lst) == 0 {
		return "", errors.New("empty list given")
	}

	return lst[rand.Intn(len(lst))], nil // #nosec:G404 // This is used to select a random string from a list, no crypto-use
}

func stableRandomFromSeed(seed string) (float64, error) {
	seedValue, err := stringToSeed(seed)
	if err != nil {
		return 0, errors.Wrap(err, "generating seed")
	}

	return rand.New(rand.NewSource(seedValue)).Float64(), nil // #nosec:G404 // Only used for generating a random number from static string, no need for cryptographic safety
}

func stringToSeed(s string) (int64, error) {
	hash := md5.New() // #nosec:G401 // Unly used to convert a string into a numer, no need for cryptographic safety
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
