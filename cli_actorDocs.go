package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/Luzifer/go_helpers/cli"
)

func init() {
	cliTool.Add(cli.RegistryEntry{
		Name:        "actor-docs",
		Description: "Generate markdown documentation for available actors",
		Run: func([]string) error {
			doc, err := generateActorDocs()
			if err != nil {
				return fmt.Errorf("generating actor docs: %w", err)
			}
			if _, err = os.Stdout.Write(append(bytes.TrimSpace(doc), '\n')); err != nil {
				return fmt.Errorf("writing actor docs to stdout: %w", err)
			}

			return nil
		},
	})
}
