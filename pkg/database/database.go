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
		// Close closes the database connector.
		Close() error

		// CopyDatabase copies all core database objects from src to target.
		CopyDatabase(src, target *gorm.DB) error

		// DB returns the underlying GORM database handle.
		DB() *gorm.DB

		// DeleteCoreMeta removes a core metadata entry by key.
		DeleteCoreMeta(key string) error

		// ReadCoreMeta reads a JSON-encoded core metadata entry by key into value.
		ReadCoreMeta(key string, value any) error

		// StoreCoreMeta stores value as a JSON-encoded core metadata entry by key.
		StoreCoreMeta(key string, value any) error

		// ReadEncryptedCoreMeta reads an encrypted core metadata entry by key into value.
		ReadEncryptedCoreMeta(key string, value any) error

		// ResetEncryptedCoreMeta removes all encrypted core metadata entries.
		ResetEncryptedCoreMeta() error

		// StoreEncryptedCoreMeta stores value as an encrypted core metadata entry by key.
		StoreEncryptedCoreMeta(key string, value any) error

		// DecryptField decrypts an encrypted database field value.
		DecryptField(string) (string, error)

		// EncryptField encrypts a database field value.
		EncryptField(string) (string, error)

		// ValidateEncryption verifies the configured encryption secret against stored metadata.
		ValidateEncryption() error
	}
)
