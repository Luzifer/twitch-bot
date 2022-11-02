package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"

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

type ActorScript struct{}

func (a ActorScript) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
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
		return attrs.MustBool("skip_cooldown_on_error", ptrBoolFalse), errors.Wrap(err, "running command")
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

func (a ActorScript) IsAsync() bool { return false }
func (a ActorScript) Name() string  { return "script" }

func (a ActorScript) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
	cmd, err := attrs.StringSlice("command")
	if err != nil || len(cmd) == 0 {
		return errors.New("command must be slice of strings with length > 0")
	}

	for i, el := range cmd {
		if err = tplValidator(el); err != nil {
			return errors.Wrapf(err, "validating cmd template (element %d)", i)
		}
	}

	return nil
}
