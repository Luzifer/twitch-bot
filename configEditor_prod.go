//go:build !dev

package main

import (
	"embed"
	"net/http"
)

var (
	//go:embed editor/*
	configEditorFrontendFS embed.FS

	configEditorFrontend = http.FS(configEditorFrontendFS)
)
