package main

import (
	"bytes"
	"os"

	"github.com/pkg/errors"
)

func init() {
	cli.Add(cliRegistryEntry{
		Name:        "actor-docs",
		Description: "Generate markdown documentation for available actors",
		Run: func(args []string) error {
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
