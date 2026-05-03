package main

import (
	"fmt"

	"github.com/Luzifer/go_helpers/cli"
)

func init() {
	cliTool.Add(cli.RegistryEntry{
		Name:        "validate-config",
		Description: "Try to load configuration file and report errors if any",
		Run: func([]string) (err error) {
			if err = loadConfig(cfg.Config); err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			return nil
		},
	})
}
