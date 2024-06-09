package main

import (
	"bytes"
	"os"

	"github.com/Luzifer/go_helpers/v2/cli"
	"github.com/pkg/errors"
)

func init() {
	cliTool.Add(cli.RegistryEntry{
		Name:        "actor-docs",
		Description: "Generate markdown documentation for available actors",
		Run: func([]string) error {
			doc, err := generateActorDocs()
			if err != nil {
				return errors.Wrap(err, "generating actor docs")
			}
			if _, err = os.Stdout.Write(append(bytes.TrimSpace(doc), '\n')); err != nil {
				return errors.Wrap(err, "writing actor docs to stdout")
			}

			return nil
		},
	})
}
