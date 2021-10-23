//go:build cgo && (linux || darwin || freebsd)

package main

import (
	"os"
	"path"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func loadPlugins(pluginDir string) error {
	logger := log.WithField("plugin_dir", pluginDir)

	d, err := os.Stat(pluginDir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Debug("Plugin directory not found, skipping")
			return nil
		}
		return errors.Wrap(err, "getting plugin-dir info")
	}

	if !d.IsDir() {
		return errors.New("plugin-dir is not a directory")
	}

	args := getRegistrationArguments()

	return errors.Wrap(filepath.Walk(pluginDir, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(currentPath, ".so") {
			// Ignore that file, is not a plugin
			return nil
		}

		logger := log.WithField("plugin", path.Base(currentPath))

		p, err := plugin.Open(currentPath)
		if err != nil {
			logger.WithError(err).Error("Unable to open plugin")
			return nil
		}

		f, err := p.Lookup("Register")
		if err != nil {
			logger.WithError(err).Error("Unable to find register function")
			return nil
		}

		f.(func(plugins.RegistrationArguments) error)(args)

		return nil
	}), "loading plugins")
}
