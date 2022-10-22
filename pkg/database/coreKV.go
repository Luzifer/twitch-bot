package database

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	coreKV struct {
		Key   string `gorm:"primaryKey"`
		Value string
	}
)

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

// StoreEncryptedCoreMeta works like StoreCoreMeta but encrypts the
// marshalled value before storing it
func (c connector) StoreEncryptedCoreMeta(key string, value any) error {
	return c.storeCoreMeta(key, value, c.EncryptField)
}

func (c connector) readCoreMeta(key string, value any, processor func(string) (string, error)) (err error) {
	var data coreKV

	if err = c.db.First(&data, "key = ?", key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCoreMetaNotFound
		}
		return errors.Wrap(err, "querying core meta table")
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

	data := coreKV{Key: key, Value: encValue}
	return errors.Wrap(
		c.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).Create(data).Error,
		"upserting core meta value",
	)
}
