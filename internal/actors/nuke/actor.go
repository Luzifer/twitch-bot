// Package nuke contains a hateraid protection actor recording messages
// in all channels for a certain period of time being able to "nuke"
// their authors by regular expression based on past messages
package nuke

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/v2/fieldcollection"
	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const (
	actorName          = "nuke"
	storeRetentionTime = 10 * time.Minute
)

var (
	botTwitchClient func() *twitch.Client
	formatMessage   plugins.MsgFormatter

	messageStore     = map[string][]*storedMessage{}
	messageStoreLock sync.RWMutex
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	botTwitchClient = args.GetTwitchClient
	formatMessage = args.FormatMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Mass ban, delete, or timeout messages based on regex. Be sure you REALLY know what you do before using this! Used wrongly this will cause a lot of damage!",
		Name:        "Nuke Chat",
		Type:        actorName,

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "10m",
				Description:     "How long to scan into the past, template must yield a duration (max 10m)",
				Key:             "scan",
				Name:            "Scan-Duration",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "delete",
				Description:     "What action to take when message matches (delete / ban / <timeout duration>)",
				Key:             "action",
				Name:            "Match-Action",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "Regular expression (RE2) to select matching messages",
				Key:             "match",
				Name:            "Message-Match",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	if _, err := args.RegisterCron("@every 1m", cleanupMessageStore); err != nil {
		return errors.Wrap(err, "registering cleanup cron")
	}

	if err := args.RegisterRawMessageHandler(rawMessageHandler); err != nil {
		return errors.Wrap(err, "registering raw message handler")
	}

	return nil
}

func cleanupMessageStore() {
	messageStoreLock.Lock()
	defer messageStoreLock.Unlock()

	var storeDeletes []string

	for ch, msgs := range messageStore {
		var idx int
		for idx = 0; idx < len(msgs); idx++ {
			if time.Since(msgs[idx].Time) < storeRetentionTime {
				break
			}
		}

		newMsgs := msgs[idx:]
		if len(newMsgs) == 0 {
			storeDeletes = append(storeDeletes, ch)
			continue
		}

		messageStore[ch] = newMsgs
		log.WithFields(log.Fields{
			"channel":         ch,
			"stored_messages": len(newMsgs),
		}).Trace("[nuke] Cleared old stored messages")
	}

	for _, ch := range storeDeletes {
		delete(messageStore, ch)
		log.WithFields(log.Fields{
			"channel": ch,
		}).Trace("[nuke] Channel is no longer stored")
	}
}

func rawMessageHandler(m *irc.Message) error {
	if m.Command != "PRIVMSG" {
		// We care only about user written messages and drop the rest
		return nil
	}

	messageStoreLock.Lock()
	defer messageStoreLock.Unlock()

	messageStore[plugins.DeriveChannel(m, nil)] = append(
		messageStore[plugins.DeriveChannel(m, nil)],
		&storedMessage{Time: time.Now(), Msg: m},
	)

	return nil
}

type (
	actor struct{}

	storedMessage struct {
		Time time.Time
		Msg  *irc.Message
	}
)

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	rawMatch, err := formatMessage(attrs.MustString("match", nil), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "formatting match")
	}
	match := regexp.MustCompile(rawMatch)

	rawScan, err := formatMessage(attrs.MustString("scan", helpers.Ptr("10m")), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "formatting scan duration")
	}
	scan, err := time.ParseDuration(rawScan)
	if err != nil {
		return false, errors.Wrap(err, "parsing scan duration")
	}
	scanTime := time.Now().Add(-scan)

	var (
		action     actionFn
		actionName string
	)
	rawAction, err := formatMessage(attrs.MustString("action", helpers.Ptr("delete")), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "formatting action")
	}
	switch rawAction {
	case "delete":
		action = actionDelete
		actionName = "delete $msgid"
	case "ban":
		action = actionBan
		actionName = "ban $user"
	default:
		to, err := time.ParseDuration(rawAction)
		if err != nil {
			return false, errors.Wrap(err, "parsing action duration")
		}
		action = getActionTimeout(to)
		actionName = "timeout $user"
	}

	channel := plugins.DeriveChannel(m, eventData)

	messageStoreLock.RLock()
	defer messageStoreLock.RUnlock()

	var executedEnforcement []string

	for _, stMsg := range messageStore[channel] {
		badges := twitch.ParseBadgeLevels(stMsg.Msg)

		if stMsg.Time.Before(scanTime) {
			continue
		}

		if badges.Has("broadcaster") || badges.Has("moderator") {
			continue
		}

		if !match.MatchString(stMsg.Msg.Trailing()) {
			continue
		}

		enforcement := strings.NewReplacer(
			"$msgid", stMsg.Msg.Tags["id"],
			"$user", plugins.DeriveUser(stMsg.Msg, nil),
		).Replace(actionName)

		if str.StringInSlice(enforcement, executedEnforcement) {
			continue
		}

		if err = action(channel, rawMatch, stMsg.Msg.Tags["id"], plugins.DeriveUser(stMsg.Msg, nil)); err != nil {
			return false, errors.Wrap(err, "executing action")
		}

		executedEnforcement = append(executedEnforcement, enforcement)
	}

	return false, nil
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "match", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "action", Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "scan", Type: fieldcollection.SchemaFieldTypeString}),
		helpers.SchemaValidateTemplateField(tplValidator, "scan", "action", "match"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}
