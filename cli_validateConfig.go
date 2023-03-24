package main

import "github.com/pkg/errors"

func init() {
	cli.Add(cliRegistryEntry{
		Name:        "validate-config",
		Description: "Try to load configuration file and report errors if any",
		Run: func(args []string) error {
			return errors.Wrap(
				loadConfig(cfg.Config),
				"loading config",
			)
		},
	})
}
