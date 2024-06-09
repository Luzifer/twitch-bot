package main

import (
	"os"

	"github.com/Luzifer/go_helpers/v2/cli"
	"github.com/gofrs/uuid/v3"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func init() {
	cliTool.Add(cli.RegistryEntry{
		Name:        "api-token",
		Description: "Generate an api-token to be entered into the config",
		Params:      []string{"<token-name>", "<scope>", "[...scope]"},
		Run: func(args []string) error {
			if len(args) < 3 { //nolint:mnd // Just a count of parameters
				return errors.New("Usage: twitch-bot api-token <token name> <scope> [...scope]")
			}

			t := configAuthToken{
				Name:    args[1],
				Modules: args[2:],
			}

			if err := fillAuthToken(&t); err != nil {
				return errors.Wrap(err, "generating token")
			}

			log.WithField("token", t.Token).Info("Token generated, add this to your config:")
			if err := yaml.NewEncoder(os.Stdout).Encode(map[string]map[string]configAuthToken{
				"auth_tokens": {
					uuid.Must(uuid.NewV4()).String(): t,
				},
			}); err != nil {
				return errors.Wrap(err, "printing token info")
			}

			return nil
		},
	})
}
