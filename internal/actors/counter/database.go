package counter

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
)

type (
	counter struct {
		Name         string `gorm:"primaryKey"`
		Value        int64
		FirstSeen    time.Time
		LastModified time.Time
	}
)

func getCounterValue(db database.Connector, counterName string) (int64, error) {
	var c counter

	err := helpers.Retry(func() error {
		err := db.DB().First(&c, "name = ?", counterName).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}

		return err
	})

	return c.Value, errors.Wrap(err, "querying counter")
}

//revive:disable-next-line:flag-parameter
func updateCounter(db database.Connector, counterName string, value int64, absolute bool, atTime time.Time) error {
	if !absolute {
		cv, err := getCounterValue(db, counterName)
		if err != nil {
			return errors.Wrap(err, "getting previous value")
		}

		value += cv
	}

	return errors.Wrap(
		helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
			return tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "name"}},
				DoUpdates: clause.AssignmentColumns([]string{"last_modified", "value"}),
			}).Create(counter{Name: counterName, Value: value, FirstSeen: atTime.UTC(), LastModified: atTime.UTC()}).Error
		}),
		"storing counter value",
	)
}

func getCounterRank(db database.Connector, prefix, name string) (rank, count int64, err error) {
	var cc []counter

	if err = helpers.Retry(func() error {
		return db.DB().
			Order("value DESC").
			Find(&cc, "name LIKE ?", prefix+"%").
			Error
	}); err != nil {
		return 0, 0, errors.Wrap(err, "querying counters")
	}

	for i, c := range cc {
		count++
		if c.Name == name {
			rank = int64(i + 1)
		}
	}

	return rank, count, nil
}

func getCounterTopList(db database.Connector, prefix string, n int, orderBy ...string) ([]counter, error) {
	var (
		cc []counter

		order          string
		validOrderCols = []string{"first_seen", "last_modified", "name", "value"}
		validOrderDirs = []string{"ASC", "DESC"}
	)

	if len(orderBy) == 0 || orderBy[0] == "" {
		order = "value DESC"
	} else {
		order = orderBy[0]
	}

	col, dir, _ := strings.Cut(order, " ")
	if col == "" {
		col = "value"
	}
	if dir == "" {
		dir = "ASC"
	}

	if !str.StringInSlice(col, validOrderCols) {
		return nil, fmt.Errorf("invalid orderBy column")
	}
	if !str.StringInSlice(dir, validOrderDirs) {
		return nil, fmt.Errorf("invalid orderBy direction")
	}

	err := helpers.Retry(func() error {
		return db.DB().
			Order(strings.Join([]string{col, dir}, " ")).
			Limit(n).
			Find(&cc, "name LIKE ?", prefix+"%").
			Error
	})

	return cc, errors.Wrap(err, "querying counters")
}
