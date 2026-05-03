package quotedb

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

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

func addQuote(db database.Connector, channel, quoteStr string) error {
	if err := helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
		return tx.Create(&quote{
			Channel:   channel,
			CreatedAt: time.Now().UnixNano(),
			Quote:     quoteStr,
		}).Error
	}); err != nil {
		return fmt.Errorf("adding quote to database: %w", err)
	}

	return nil
}

func delQuote(db database.Connector, channel string, quoteIdx int) error {
	_, createdAt, _, err := getQuoteRaw(db, channel, quoteIdx)
	if err != nil {
		return fmt.Errorf("fetching specified quote: %w", err)
	}

	if err = helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
		return tx.Delete(&quote{}, "channel = ? AND created_at = ?", channel, createdAt).Error
	}); err != nil {
		return fmt.Errorf("deleting quote: %w", err)
	}

	return nil
}

func getChannelQuotes(db database.Connector, channel string) ([]string, error) {
	var qs []quote
	if err := helpers.Retry(func() error {
		return db.DB().Where("channel = ?", channel).Order("created_at").Find(&qs).Error
	}); err != nil {
		return nil, fmt.Errorf("querying quotes: %w", err)
	}

	var quotes []string
	for _, q := range qs {
		quotes = append(quotes, q.Quote)
	}

	return quotes, nil
}

func getMaxQuoteIdx(db database.Connector, channel string) (int, error) {
	var count int64
	if err := helpers.Retry(func() error {
		return db.DB().
			Model(&quote{}).
			Where("channel = ?", channel).
			Count(&count).
			Error
	}); err != nil {
		return 0, fmt.Errorf("getting quote count: %w", err)
	}

	return int(count), nil
}

func getQuote(db database.Connector, channel string, quote int) (int, string, error) {
	quoteIdx, _, quoteText, err := getQuoteRaw(db, channel, quote)
	return quoteIdx, quoteText, err
}

func getQuoteRaw(db database.Connector, channel string, quoteIdx int) (int, int64, string, error) {
	if quoteIdx == 0 {
		maxQuoteIdx, err := getMaxQuoteIdx(db, channel)
		if err != nil {
			return 0, 0, "", fmt.Errorf("getting max quote idx: %w", err)
		}
		quoteIdx = rand.Intn(maxQuoteIdx) + 1 // #nosec G404 // no need for cryptographic safety
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
		return 0, 0, "", fmt.Errorf("getting quote from DB: %w", err)
	}
}

func setQuotes(db database.Connector, channel string, quotes []string) error {
	if err := helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
		if err := tx.Where("channel = ?", channel).Delete(&quote{}).Error; err != nil {
			return fmt.Errorf("deleting quotes for channel: %w", err)
		}

		t := time.Now()
		for _, quoteStr := range quotes {
			if err := tx.Create(&quote{
				Channel:   channel,
				CreatedAt: t.UnixNano(),
				Quote:     quoteStr,
			}).Error; err != nil {
				return fmt.Errorf("adding quote: %w", err)
			}

			t = t.Add(time.Nanosecond) // Increase by one ns to adhere to unique index
		}

		return nil
	}); err != nil {
		return fmt.Errorf("replacing quotes: %w", err)
	}

	return nil
}

func updateQuote(db database.Connector, channel string, idx int, quoteStr string) error {
	_, createdAt, _, err := getQuoteRaw(db, channel, idx)
	if err != nil {
		return fmt.Errorf("fetching specified quote: %w", err)
	}

	if err = helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
		return tx.Where("channel = ? AND created_at = ?", channel, createdAt).
			Update("quote", quoteStr).
			Error
	}); err != nil {
		return fmt.Errorf("updating quote: %w", err)
	}

	return nil
}
