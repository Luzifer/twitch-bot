// Package filesay contains an actor to paste a remote URL as chat
// commands i.e. for bulk banning users
package filesay

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const (
	actorName = "filesay"

	httpTimeout = 5 * time.Second
)

type actor struct{}

var (
	formatMessage plugins.MsgFormatter
	send          plugins.SendMessageFunc
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage
	send = args.SendMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Takes the content of an URL and pastes it to the current channel",
		Name:        "FileSay",
		Type:        actorName,

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Source of the content to post",
				Key:             "source",
				Name:            "source",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	return nil
}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	ptrStringEmpty := func(v string) *string { return &v }("")

	source, err := formatMessage(attrs.MustString("source", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, fmt.Errorf("executing source template: %w", err)
	}

	if source == "" {
		return false, errors.New("source template evaluated to empty string")
	}

	if _, err := url.Parse(source); err != nil {
		return false, fmt.Errorf("parsing URL: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return false, fmt.Errorf("creating HTTP request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("executing HTTP request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing response body (leaked fd)")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("http status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if err = send(&irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				plugins.DeriveChannel(m, eventData),
				scanner.Text(),
			},
		}); err != nil {
			return false, fmt.Errorf("sending message: %w", err)
		}
	}

	return false, nil
}

func (actor) IsAsync() bool { return true }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "source", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "source"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}
