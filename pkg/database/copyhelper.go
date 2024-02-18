package database

import (
	"reflect"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const copyBatchSize = 100

// CopyObjects is a helper to copy elements of a given type from the
// src to the target GORM database interface
func CopyObjects(src, target *gorm.DB, objects ...any) (err error) {
	for _, obj := range objects {
		copySlice := reflect.New(reflect.SliceOf(reflect.TypeOf(obj))).Elem().Addr().Interface()

		if err = target.AutoMigrate(obj); err != nil {
			return errors.Wrap(err, "applying migration to target")
		}

		if err = target.Where("1 = 1").Delete(obj).Error; err != nil {
			return errors.Wrap(err, "cleaning target table")
		}

		if err = src.FindInBatches(copySlice, copyBatchSize, func(*gorm.DB, int) error {
			if err = target.Save(copySlice).Error; err != nil {
				if errors.Is(err, gorm.ErrEmptySlice) {
					// That's fine and no reason to exit here
					return nil
				}
				return errors.Wrap(err, "inserting collected elements")
			}

			return nil
		}).Error; err != nil {
			return errors.Wrap(err, "batch-copying data")
		}
	}

	return nil
}
