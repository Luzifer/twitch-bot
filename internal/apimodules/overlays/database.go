package overlays

import (
	"bytes"
	"embed"
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/plugins"
)

//go:embed schema/**
var schema embed.FS

func addEvent(channel string, evt socketMessage) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(evt.Fields); err != nil {
		return errors.Wrap(err, "encoding fields")
	}

	_, err := db.DB().Exec(
		`INSERT INTO overlays_events
			(channel, created_at, event_type, fields)
			VALUES ($1, $2, $3, $4);`,
		channel, evt.Time.UnixNano(), evt.Type, buf.String(),
	)

	return errors.Wrap(err, "storing event to database")
}

func getChannelEvents(channel string) ([]socketMessage, error) {
	rows, err := db.DB().Query(
		`SELECT created_at, event_type, fields
			FROM overlays_events
			WHERE channel = $1
			ORDER BY created_at;`,
		channel,
	)
	if err != nil {
		return nil, errors.Wrap(err, "querying channel events")
	}

	var out []socketMessage
	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, errors.Wrap(err, "advancing row read")
		}

		var (
			createdAt            int64
			eventType, rawFields string
		)
		if err = rows.Scan(&createdAt, &eventType, &rawFields); err != nil {
			return nil, errors.Wrap(err, "scanning row")
		}

		fields := new(plugins.FieldCollection)
		if err = json.NewDecoder(strings.NewReader(rawFields)).Decode(fields); err != nil {
			return nil, errors.Wrap(err, "decoding fields")
		}

		out = append(out, socketMessage{
			IsLive: false,
			Time:   time.Unix(0, createdAt),
			Type:   eventType,
			Fields: fields,
		})
	}

	return out, errors.Wrap(rows.Err(), "advancing row read")
}
