package main

import (
	"bytes"
	"os"

	"github.com/pkg/errors"
)

func init() {
	cli.Add(cliRegistryEntry{
		Name:        "tpl-docs",
		Description: "Generate markdown documentation for available template functions",
		Run: func(args []string) error {
			doc, err := generateTplDocs()
			if err != nil {
				return errors.Wrap(err, "generating template docs")
			}
			if _, err = os.Stdout.Write(append(bytes.TrimSpace(doc), '\n')); err != nil {
				return errors.Wrap(err, "writing actor docs to stdout")
			}

			return nil
		},
	})
}
