package overlays

import (
	"io/fs"
	"net/http"
	"path"
)

// Compile-time assertion
var _ http.FileSystem = httpFSStack{}

type httpFSStack []http.FileSystem

func (h httpFSStack) Open(name string) (http.File, error) {
	for _, fs := range h {
		if f, err := fs.Open(name); err == nil {
			return f, nil
		}
	}

	return nil, fs.ErrNotExist
}

// Compile-time assertion
var _ http.FileSystem = prefixedFS{}

type prefixedFS struct {
	originFS http.FileSystem
	prefix   string
}

func newPrefixedFS(prefix string, originFS http.FileSystem) *prefixedFS {
	return &prefixedFS{originFS: originFS, prefix: prefix}
}

func (p prefixedFS) Open(name string) (http.File, error) {
	return p.originFS.Open(path.Join(p.prefix, name))
}
