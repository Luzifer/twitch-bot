package overlays

import (
	"io/fs"
	"net/http"
	"path"
)

type (
	httpFSStack []http.FileSystem

	prefixedFS struct {
		originFS http.FileSystem
		prefix   string
	}
)

// Compile-time assertion
var (
	_ http.FileSystem = httpFSStack{} //revive:disable-line:enforce-slice-style // needed for compile-time assertion
	_ http.FileSystem = prefixedFS{}
)

func (h httpFSStack) Open(name string) (http.File, error) {
	for _, stackedFS := range h {
		if f, err := stackedFS.Open(name); err == nil {
			return f, nil
		}
	}

	return nil, fs.ErrNotExist
}

func newPrefixedFS(prefix string, originFS http.FileSystem) *prefixedFS {
	return &prefixedFS{originFS: originFS, prefix: prefix}
}

func (p prefixedFS) Open(name string) (http.File, error) {
	return p.originFS.Open(path.Join(p.prefix, name)) //nolint:wrapcheck // pass through original error, we're just a thin wrapper
}
