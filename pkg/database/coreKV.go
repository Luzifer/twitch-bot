package database

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Luzifer/go_helpers/v2/backoff"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
)

const (
	encryptionValidationKey        = "encryption-validation"
	encryptionValidationMinBackoff = 500 * time.Millisecond
	encryptionValidationTries      = 5
)

type (
	coreKV struct {
		Name  string `gorm:"primaryKey"`
		Value string
	}
)

// DeleteCoreMeta removes a core_kv table entry
func (c connector) DeleteCoreMeta(key string) error {
	return errors.Wrap(
		helpers.RetryTransaction(c.db, func(tx *gorm.DB) error {
			return tx.Delete(&coreKV{}, "name = ?", key).Error
		}),
		"deleting key from database",
	)
}

// ReadCoreMeta reads an entry of the core_kv table specified by
// the given `key` and unmarshals it into the `value`. The value must
// be a valid variable to `json.NewDecoder(...).Decode(value)`
// (pointer to struct, string, int, ...). In case the key does not
// exist a check to 'errors.Is(err, sql.ErrNoRows)' will succeed
func (c connector) ReadCoreMeta(key string, value any) error {
	return c.readCoreMeta(key, value, nil)
}

// StoreCoreMeta stores an entry to the core_kv table soecified by
// the given `key`. The value given must be a valid variable to
// `json.NewEncoder(...).Encode(value)`.
func (c connector) StoreCoreMeta(key string, value any) error {
	return c.storeCoreMeta(key, value, nil)
}

// ReadEncryptedCoreMeta works like ReadCoreMeta but decrypts the
// stored value before unmarshalling it
func (c connector) ReadEncryptedCoreMeta(key string, value any) error {
	return c.readCoreMeta(key, value, c.DecryptField)
}

// ResetEncryptedCoreMeta removes all CoreKV entries from the database
func (c connector) ResetEncryptedCoreMeta() error {
	return errors.Wrap(
		helpers.RetryTransaction(c.db, func(tx *gorm.DB) error {
			return tx.Delete(&coreKV{}, "value LIKE ?", "U2FsdGVkX1%").Error
		}),
		"removing encrypted meta entries",
	)
}

// StoreEncryptedCoreMeta works like StoreCoreMeta but encrypts the
// marshalled value before storing it
func (c connector) StoreEncryptedCoreMeta(key string, value any) error {
	return c.storeCoreMeta(key, value, c.EncryptField)
}

func (c connector) ValidateEncryption() error {
	var (
		storedHash     string
		validationHash = fmt.Sprintf("%x", sha512.Sum512([]byte(c.encryptionSecret)))
	)

	err := backoff.NewBackoff().
		WithMaxIterations(encryptionValidationTries).
		WithMinIterationTime(encryptionValidationMinBackoff).
		Retry(func() error {
			return c.ReadEncryptedCoreMeta(encryptionValidationKey, &storedHash)
		})

	switch {
	case err == nil:
		if storedHash != validationHash {
			// Shouldn't happen: When decryption is possible it should match
			return errors.New("mismatch between expected and stored hash")
		}
		return nil

	case errors.Is(err, ErrCoreMetaNotFound):
		return errors.Wrap(
			c.StoreEncryptedCoreMeta(encryptionValidationKey, validationHash),
			"initializing encryption validation",
		)

	default:
		return errors.Wrap(err, "reading encryption-validation")
	}
}

//revive:disable-next-line:confusing-naming
func (c connector) readCoreMeta(key string, value any, processor func(string) (string, error)) (err error) {
	var data coreKV

	if err = helpers.Retry(func() error {
		err = c.db.First(&data, "name = ?", key).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return backoff.NewErrCannotRetry(ErrCoreMetaNotFound)
		}
		return errors.Wrap(err, "querying core meta table")
	}); err != nil {
		return err
	}

	if data.Value == "" {
		return errors.New("empty value returned")
	}

	if processor != nil {
		if data.Value, err = processor(data.Value); err != nil {
			return errors.Wrap(err, "processing stored value")
		}
	}

	if err := json.NewDecoder(strings.NewReader(data.Value)).Decode(value); err != nil {
		return errors.Wrap(err, "JSON decoding value")
	}

	return nil
}

//revive:disable-next-line:confusing-naming
func (c connector) storeCoreMeta(key string, value any, processor func(string) (string, error)) (err error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(value); err != nil {
		return errors.Wrap(err, "JSON encoding value")
	}

	encValue := strings.TrimSpace(buf.String())
	if processor != nil {
		if encValue, err = processor(encValue); err != nil {
			return errors.Wrap(err, "processing value to store")
		}
	}

	data := coreKV{Name: key, Value: encValue}
	return errors.Wrap(
		helpers.RetryTransaction(c.db, func(tx *gorm.DB) error {
			return tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "name"}},
				DoUpdates: clause.AssignmentColumns([]string{"value"}),
			}).Create(data).Error
		}),
		"upserting core meta value",
	)
}
