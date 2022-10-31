package filesay

import (
	"bufio"
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v2/plugins"
)

const (
	actorName = "filesay"

	httpTimeout = 5 * time.Second
)

var (
	formatMessage plugins.MsgFormatter
	send          plugins.SendMessageFunc
)

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

type actor struct{}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	ptrStringEmpty := func(v string) *string { return &v }("")

	source, err := formatMessage(attrs.MustString("source", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing source template")
	}

	if source == "" {
		return false, errors.New("source template evaluated to empty string")
	}

	if _, err := url.Parse(source); err != nil {
		return false, errors.Wrap(err, "parsing URL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return false, errors.Wrap(err, "creating HTTP request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, errors.Wrap(err, "executing HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, errors.Errorf("http status %d", resp.StatusCode)
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
			return false, errors.Wrap(err, "sending message")
		}
	}

	return false, nil
}

func (a actor) IsAsync() bool { return true }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) error {
	sourceTpl, err := attrs.String("source")
	if err != nil || sourceTpl == "" {
		return errors.New("source is expected to be non-empty string")
	}

	if err = tplValidator(sourceTpl); err != nil {
		return errors.Wrap(err, "validating source template")
	}

	return nil
}
