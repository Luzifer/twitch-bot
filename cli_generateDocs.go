//go:build docgen

package main

import (
	"fmt"

	"github.com/Luzifer/go_helpers/cli"

	"github.com/Luzifer/twitch-bot/v3/internal/docgen"
)

func init() {
	cliTool.Add(cli.RegistryEntry{
		Name:        "generate-docs",
		Description: "Generate project documentation",
		Run: func([]string) error {
			if err := docgen.GenerateActorDocs(availableActorDocs); err != nil {
				return fmt.Errorf("generating actor docs: %w", err)
			}

			if err := docgen.GenerateTplDocs(tplFuncs.docs, formatMessage); err != nil {
				return fmt.Errorf("generating template docs: %w", err)
			}

			return nil
		},
	})
}
