package quotedb

import (
	"database/sql"
	"embed"
	"math/rand"
	"time"

	"github.com/pkg/errors"
)

//go:embed schema/**
var schema embed.FS

func addQuote(channel, quote string) error {
	_, err := db.DB().Exec(
		`INSERT INTO quotedb
			(channel, created_at, quote)
			VALUES ($1, $2, $3);`,
		channel, time.Now().UnixNano(), quote,
	)

	return errors.Wrap(err, "adding quote to database")
}

func delQuote(channel string, quote int) error {
	_, createdAt, _, err := getQuoteRaw(channel, quote)
	if err != nil {
		return errors.Wrap(err, "fetching specified quote")
	}

	_, err = db.DB().Exec(
		`DELETE FROM quotedb
			WHERE channel = $1 AND created_at = $2;`,
		channel, createdAt,
	)

	return errors.Wrap(err, "deleting quote")
}

func getChannelQuotes(channel string) ([]string, error) {
	rows, err := db.DB().Query(
		`SELECT quote
			FROM quotedb
			WHERE channel = $1
			ORDER BY created_at ASC`,
		channel,
	)
	if err != nil {
		return nil, errors.Wrap(err, "querying quotes")
	}

	var quotes []string
	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, errors.Wrap(err, "advancing row read")
		}

		var quote string
		if err = rows.Scan(&quote); err != nil {
			return nil, errors.Wrap(err, "scanning row")
		}

		quotes = append(quotes, quote)
	}

	return quotes, errors.Wrap(rows.Err(), "advancing row read")
}

func getMaxQuoteIdx(channel string) (int, error) {
	row := db.DB().QueryRow(
		`SELECT COUNT(1) as quoteCount
			FROM quotedb
			WHERE channel = $1;`,
		channel,
	)

	var count int
	err := row.Scan(&count)

	return count, errors.Wrap(err, "getting quote count")
}

func getQuote(channel string, quote int) (int, string, error) {
	quoteIdx, _, quoteText, err := getQuoteRaw(channel, quote)
	return quoteIdx, quoteText, err
}

func getQuoteRaw(channel string, quote int) (int, int64, string, error) {
	if quote == 0 {
		max, err := getMaxQuoteIdx(channel)
		if err != nil {
			return 0, 0, "", errors.Wrap(err, "getting max quote idx")
		}
		quote = rand.Intn(max) + 1 // #nosec G404 // no need for cryptographic safety
	}

	row := db.DB().QueryRow(
		`SELECT created_at, quote
			FROM quotedb
			WHERE channel = $1
			ORDER BY created_at ASC
			LIMIT 1 OFFSET $2`,
		channel, quote-1,
	)

	var (
		createdAt int64
		quoteText string
	)

	err := row.Scan(&createdAt, &quoteText)
	switch {
	case err == nil:
		return quote, createdAt, quoteText, nil

	case errors.Is(err, sql.ErrNoRows):
		return 0, 0, "", nil

	default:
		return 0, 0, "", errors.Wrap(err, "getting quote from DB")
	}
}

func setQuotes(channel string, quotes []string) error {
	tx, err := db.DB().Begin()
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}

	if _, err = tx.Exec(
		`DELETE FROM quotedb
			WHERE channel = $1;`,
		channel,
	); err != nil {
		defer tx.Rollback()
		return errors.Wrap(err, "deleting quotes for channel")
	}

	t := time.Now()
	for _, quote := range quotes {
		if _, err = tx.Exec(
			`INSERT INTO quotedb
				(channel, created_at, quote)
				VALUES ($1, $2, $3);`,
			channel, t.UnixNano(), quote,
		); err != nil {
			defer tx.Rollback()
			return errors.Wrap(err, "adding quote for channel")
		}

		t = t.Add(time.Nanosecond) // Increase by one ns to adhere to unique index
	}

	return errors.Wrap(tx.Commit(), "committing change")
}

func updateQuote(channel string, idx int, quote string) error {
	_, createdAt, _, err := getQuoteRaw(channel, idx)
	if err != nil {
		return errors.Wrap(err, "fetching specified quote")
	}

	_, err = db.DB().Exec(
		`UPDATE quotedb
			SET quote = $3
			WHERE channel = $1
				AND created_at = $2;`,
		channel, createdAt, quote,
	)

	return errors.Wrap(err, "updating quote")
}
