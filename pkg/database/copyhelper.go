package database

import (
	"errors"
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

const copyBatchSize = 100

// CopyObjects is a helper to copy elements of a given type from the
// src to the target GORM database interface
func CopyObjects(src, target *gorm.DB, objects ...any) (err error) {
	for _, obj := range objects {
		copySlice := reflect.New(reflect.SliceOf(reflect.TypeOf(obj))).Elem().Addr().Interface()

		if err = target.AutoMigrate(obj); err != nil {
			return fmt.Errorf("applying migration to target: %w", err)
		}

		if err = target.Where("1 = 1").Delete(obj).Error; err != nil {
			return fmt.Errorf("cleaning target table: %w", err)
		}

		if err = src.FindInBatches(copySlice, copyBatchSize, func(*gorm.DB, int) error {
			if err = target.Save(copySlice).Error; err != nil {
				if errors.Is(err, gorm.ErrEmptySlice) {
					// That's fine and no reason to exit here
					return nil
				}
				return fmt.Errorf("inserting collected elements: %w", err)
			}

			return nil
		}).Error; err != nil {
			return fmt.Errorf("batch-copying data: %w", err)
		}
	}

	return nil
}
