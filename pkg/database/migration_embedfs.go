package database

import (
	"embed"
	"io/fs"
	"path"
	"strings"
)

type (
	// EmbedFSMigrator is a wrapper around embed.FS enabling ReadDir("/")
	// which normally would cause an error as path "/" is not available
	// within an embed.FS
	EmbedFSMigrator struct {
		BasePath string
		embed.FS
	}
)

// NewEmbedFSMigrator creates a new EmbedFSMigrator
func NewEmbedFSMigrator(fs embed.FS, basePath string) MigrationStorage {
	return EmbedFSMigrator{BasePath: basePath, FS: fs}
}

// ReadDir Wraps embed.FS.ReadDir with adjustment of the path prefix
func (e EmbedFSMigrator) ReadDir(name string) ([]fs.DirEntry, error) {
	name = path.Join(e.BasePath, strings.TrimPrefix(name, "/"))
	return e.FS.ReadDir(name)
}

// ReadFile Wraps embed.FS.ReadFile with adjustment of the path prefix
func (e EmbedFSMigrator) ReadFile(name string) ([]byte, error) {
	name = path.Join(e.BasePath, strings.TrimPrefix(name, "/"))
	return e.FS.ReadFile(name)
}
