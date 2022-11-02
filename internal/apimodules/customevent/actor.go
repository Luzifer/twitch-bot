package customevent

import (
	"strings"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

type actor struct{}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	fd, err := formatMessage(attrs.MustString("fields", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing fields template")
	}

	if fd == "" {
		return false, errors.New("fields template evaluated to empty string")
	}

	delayRaw, err := formatMessage(attrs.MustString("schedule_in", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing schedule_in template")
	}

	if delay, err := time.ParseDuration(delayRaw); err == nil && delay > 0 {
		fields, err := parseEvent(plugins.DeriveChannel(m, eventData), strings.NewReader(fd))
		if err != nil {
			return false, errors.Wrap(err, "parsing fields data")
		}

		if err = storeEvent(db, time.Now().Add(delay).UTC(), plugins.DeriveChannel(m, eventData), fields); err != nil {
			return false, errors.Wrap(err, "storing event")
		}

		return false, errors.Wrap(mc.Refresh(), "refreshing memory cache")
	}

	return false, errors.Wrap(
		triggerEvent(plugins.DeriveChannel(m, eventData), strings.NewReader(fd)),
		"triggering event",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
	if v, err := attrs.String("fields"); err != nil || v == "" {
		return errors.New("fields is expected to be non-empty string")
	}

	for _, field := range []string{"fields", "schedule_in"} {
		if err = tplValidator(attrs.MustString(field, ptrStringEmpty)); err != nil {
			return errors.Wrapf(err, "validating %s template", field)
		}
	}

	return nil
}
