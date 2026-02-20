package main

import (
	"github.com/Luzifer/go_helpers/cli"
	"github.com/pkg/errors"
)

func init() {
	cliTool.Add(cli.RegistryEntry{
		Name:        "validate-config",
		Description: "Try to load configuration file and report errors if any",
		Run: func([]string) error {
			return errors.Wrap(
				loadConfig(cfg.Config),
				"loading config",
			)
		},
	})
}
