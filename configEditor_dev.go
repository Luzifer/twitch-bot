//go:build dev

package main

import "net/http"

var configEditorFrontend http.FileSystem = http.Dir(".")
