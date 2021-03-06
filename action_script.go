package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func() Actor { return &ActorScript{} })
}

type ActorScript struct {
	Command []string `json:"command" yaml:"command"`
}

func (a ActorScript) Execute(c *irc.Client, m *irc.Message, r *Rule) error {
	if len(a.Command) == 0 {
		return nil
	}

	var command []string
	for _, arg := range a.Command {
		tmp, err := formatMessage(arg, m, r, nil)
		if err != nil {
			return errors.Wrap(err, "execute command argument template")
		}

		command = append(command, tmp)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.CommandTimeout)
	defer cancel()

	var (
		stdin  = new(bytes.Buffer)
		stdout = new(bytes.Buffer)
	)

	if err := json.NewEncoder(stdin).Encode(map[string]interface{}{
		"badges":   ircHandler{}.ParseBadgeLevels(m),
		"channel":  m.Params[0],
		"message":  m.Trailing(),
		"tags":     m.Tags,
		"username": m.User,
	}); err != nil {
		return errors.Wrap(err, "encoding script input")
	}

	cmd := exec.CommandContext(ctx, command[0], command[1:]...) // #nosec G204 // This is expected to call a command with parameters
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdin = stdin
	cmd.Stdout = stdout

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "running command")
	}

	if stdout.Len() == 0 {
		// Script was successful but did not yield actions
		return nil
	}

	var (
		actions []*RuleAction
		decoder = json.NewDecoder(stdout)
	)

	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&actions); err != nil {
		return errors.Wrap(err, "decoding actions output")
	}

	for _, action := range actions {
		if err := triggerActions(c, m, r, action); err != nil {
			return errors.Wrap(err, "execute returned action")
		}
	}

	return nil
}

func (a ActorScript) IsAsync() bool { return false }
func (a ActorScript) Name() string  { return "script" }
