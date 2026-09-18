//go:build docgen

package main

import (
	"fmt"

	"github.com/Luzifer/go_helpers/cli"

	"github.com/Luzifer/twitch-bot/v3/internal/docgen"
	"github.com/Luzifer/twitch-bot/v3/pkg/event"
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

			if err := docgen.GenerateEventDocs(event.KnownEvents()); err != nil {
				return fmt.Errorf("generating event docs: %w", err)
			}

			if err := docgen.GenerateEventTypes(event.KnownEvents()); err != nil {
				return fmt.Errorf("generating event types: %w", err)
			}

			return nil
		},
	})
}
