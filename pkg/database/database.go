// Package database represents a connector to the sqlite storage
// backend to store persistent data from core and plugins
package database

import (
	"gorm.io/gorm"
)

type (
	// Connector represents a database connection having some extra
	// convenience methods
	Connector interface {
		Close() error
		CopyDatabase(src, target *gorm.DB) error
		DB() *gorm.DB
		DeleteCoreMeta(key string) error
		ReadCoreMeta(key string, value any) error
		StoreCoreMeta(key string, value any) error
		ReadEncryptedCoreMeta(key string, value any) error
		ResetEncryptedCoreMeta() error
		StoreEncryptedCoreMeta(key string, value any) error
		DecryptField(string) (string, error)
		EncryptField(string) (string, error)
		ValidateEncryption() error
	}
)
