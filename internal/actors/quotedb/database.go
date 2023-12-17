package quotedb

import (
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
)

type (
	quote struct {
		ID        uint64 `gorm:"primaryKey"`
		Channel   string `gorm:"not null;uniqueIndex:ensure_sort_idx;size:32"`
		CreatedAt int64  `gorm:"uniqueIndex:ensure_sort_idx"`
		Quote     string
	}
)

func AddQuote(db database.Connector, channel, quoteStr string) error {
	return errors.Wrap(
		helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
			return tx.Create(&quote{
				Channel:   channel,
				CreatedAt: time.Now().UnixNano(),
				Quote:     quoteStr,
			}).Error
		}),
		"adding quote to database",
	)
}

func DelQuote(db database.Connector, channel string, quoteIdx int) error {
	_, createdAt, _, err := GetQuoteRaw(db, channel, quoteIdx)
	if err != nil {
		return errors.Wrap(err, "fetching specified quote")
	}

	return errors.Wrap(
		helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
			return tx.Delete(&quote{}, "channel = ? AND created_at = ?", channel, createdAt).Error
		}),
		"deleting quote",
	)
}

func GetChannelQuotes(db database.Connector, channel string) ([]string, error) {
	var qs []quote
	if err := helpers.Retry(func() error {
		return db.DB().Where("channel = ?", channel).Order("created_at").Find(&qs).Error
	}); err != nil {
		return nil, errors.Wrap(err, "querying quotes")
	}

	var quotes []string
	for _, q := range qs {
		quotes = append(quotes, q.Quote)
	}

	return quotes, nil
}

func GetMaxQuoteIdx(db database.Connector, channel string) (int, error) {
	var count int64
	if err := helpers.Retry(func() error {
		return db.DB().
			Model(&quote{}).
			Where("channel = ?", channel).
			Count(&count).
			Error
	}); err != nil {
		return 0, errors.Wrap(err, "getting quote count")
	}

	return int(count), nil
}

func GetQuote(db database.Connector, channel string, quote int) (int, string, error) {
	quoteIdx, _, quoteText, err := GetQuoteRaw(db, channel, quote)
	return quoteIdx, quoteText, err
}

func GetQuoteRaw(db database.Connector, channel string, quoteIdx int) (int, int64, string, error) {
	if quoteIdx == 0 {
		max, err := GetMaxQuoteIdx(db, channel)
		if err != nil {
			return 0, 0, "", errors.Wrap(err, "getting max quote idx")
		}
		quoteIdx = rand.Intn(max) + 1 // #nosec G404 // no need for cryptographic safety
	}

	var q quote
	err := helpers.Retry(func() error {
		return db.DB().
			Where("channel = ?", channel).
			Limit(1).
			Offset(quoteIdx - 1).
			First(&q).Error
	})

	switch {
	case err == nil:
		return quoteIdx, q.CreatedAt, q.Quote, nil

	case errors.Is(err, gorm.ErrRecordNotFound):
		return 0, 0, "", nil

	default:
		return 0, 0, "", errors.Wrap(err, "getting quote from DB")
	}
}

func SetQuotes(db database.Connector, channel string, quotes []string) error {
	return errors.Wrap(
		helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
			if err := tx.Where("channel = ?", channel).Delete(&quote{}).Error; err != nil {
				return errors.Wrap(err, "deleting quotes for channel")
			}

			t := time.Now()
			for _, quoteStr := range quotes {
				if err := tx.Create(&quote{
					Channel:   channel,
					CreatedAt: t.UnixNano(),
					Quote:     quoteStr,
				}).Error; err != nil {
					return errors.Wrap(err, "adding quote")
				}

				t = t.Add(time.Nanosecond) // Increase by one ns to adhere to unique index
			}

			return nil
		}),
		"replacing quotes",
	)
}

func UpdateQuote(db database.Connector, channel string, idx int, quoteStr string) error {
	_, createdAt, _, err := GetQuoteRaw(db, channel, idx)
	if err != nil {
		return errors.Wrap(err, "fetching specified quote")
	}

	return errors.Wrap(
		helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
			return tx.Where("channel = ? AND created_at = ?", channel, createdAt).
				Update("quote", quoteStr).
				Error
		}),
		"updating quote",
	)
}
