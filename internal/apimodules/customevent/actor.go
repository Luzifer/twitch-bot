package customevent

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

type actor struct{}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	fd, err := formatMessage(attrs.MustString("fields", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, fmt.Errorf("executing fields template: %w", err)
	}

	if fd == "" {
		return false, errors.New("fields template evaluated to empty string")
	}

	delayRaw, err := formatMessage(attrs.MustString("schedule_in", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, fmt.Errorf("executing schedule_in template: %w", err)
	}

	if err = triggerOrStoreEvent(plugins.DeriveChannel(m, eventData), strings.NewReader(fd), delayRaw); err != nil {
		return false, fmt.Errorf("triggering event: %w", err)
	}

	return false, nil
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "fields", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "schedule_in", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "fields", "schedule_in"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}
