package helpers

import (
	"github.com/Luzifer/go_helpers/v2/backoff"
	"gorm.io/gorm"
)

const (
	maxRetries = 5
)

// Retry contains a standard set of configuration parameters for an
// exponential backoff to be used throughout the bot
func Retry(fn func() error) error {
	//nolint:wrapcheck
	return backoff.NewBackoff().
		WithMaxIterations(maxRetries).
		Retry(fn)
}

// RetryTransaction takes a database object and a function acting on
// the database. The function will be run in a transaction on the
// database and will be retried as if executed using Retry
func RetryTransaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return Retry(func() error {
		return db.Transaction(fn) //nolint:wrapcheck
	})
}
