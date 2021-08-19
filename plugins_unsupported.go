//go:build !cgo || !(linux || darwin || freebsd)

package main

import log "github.com/sirupsen/logrus"

func loadPlugins(string) error {
	log.Warn("Plugin support is disabled in this version")
	return nil
}
