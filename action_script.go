package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

func init() {
	registerAction("script", func() plugins.Actor { return &ActorScript{} })

	registerActorDocumentation(plugins.ActionDocumentation{
		Description: "Execute external script / command",
		Name:        "Execute Script / Command",
		Type:        "script",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Command to execute",
				Key:             "command",
				Name:            "Command",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeStringSlice,
			},
			{
				Default:         "false",
				Description:     "Do not activate cooldown for route when command exits non-zero",
				Key:             "skip_cooldown_on_error",
				Name:            "Skip Cooldown on Error",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeBool,
			},
		},
	})
}

// ActorScript contains an actor to execute arbitrary commands and scripts
type ActorScript struct{}

// Execute implements actor interface
func (ActorScript) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	command, err := attrs.StringSlice("command")
	if err != nil {
		return false, errors.Wrap(err, "getting command")
	}

	for i := range command {
		tmp, err := formatMessage(command[i], m, r, eventData)
		if err != nil {
			return false, errors.Wrap(err, "execute command argument template")
		}

		command[i] = tmp
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.CommandTimeout)
	defer cancel()

	var (
		stdin  = new(bytes.Buffer)
		stdout = new(bytes.Buffer)
	)

	scriptInput := map[string]interface{}{
		"badges":   twitch.ParseBadgeLevels(m),
		"channel":  plugins.DeriveChannel(m, eventData),
		"username": plugins.DeriveUser(m, eventData),
	}

	if m != nil {
		scriptInput["message"] = m.Trailing()
		scriptInput["tags"] = m.Tags
	}

	if err := json.NewEncoder(stdin).Encode(scriptInput); err != nil {
		return false, errors.Wrap(err, "encoding script input")
	}

	cmd := exec.CommandContext(ctx, command[0], command[1:]...) // #nosec G204 // This is expected to call a command with parameters
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdin = stdin
	cmd.Stdout = stdout

	if err := cmd.Run(); err != nil {
		return attrs.MustBool("skip_cooldown_on_error", ptrBoolFalse), errors.Wrapf(err, "running command in rule %s", r.UUID)
	}

	if stdout.Len() == 0 {
		// Script was successful but did not yield actions
		return false, nil
	}

	var (
		actions []*plugins.RuleAction
		decoder = json.NewDecoder(stdout)
	)

	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&actions); err != nil {
		return false, errors.Wrap(err, "decoding actions output")
	}

	for _, action := range actions {
		apc, err := triggerAction(c, m, r, action, eventData)
		if err != nil {
			return preventCooldown, errors.Wrap(err, "execute returned action")
		}
		preventCooldown = preventCooldown || apc
	}

	return preventCooldown, nil
}

// IsAsync implements actor interface
func (ActorScript) IsAsync() bool { return false }

// Name implements actor interface
func (ActorScript) Name() string { return "script" }

// Validate implements actor interface
func (ActorScript) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "command", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeStringSlice}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "skip_cooldown_on_error", Type: fieldcollection.SchemaFieldTypeBool}),
		fieldcollection.MustHaveNoUnknowFields,
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	for i, el := range attrs.MustStringSlice("command", nil) {
		if err = tplValidator(el); err != nil {
			return errors.Wrapf(err, "validating cmd template (element %d)", i)
		}
	}

	return nil
}
