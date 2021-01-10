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
	registerAction(func(c *irc.Client, m *irc.Message, ruleDef *rule, r *ruleAction) error {
		if len(r.Command) == 0 {
			return nil
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

		cmd := exec.CommandContext(ctx, r.Command[0], r.Command[1:]...)
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
			actions []*ruleAction
			decoder = json.NewDecoder(stdout)
		)

		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&actions); err != nil {
			return errors.Wrap(err, "decoding actions output")
		}

		for _, action := range actions {
			if err := triggerActions(c, m, ruleDef, action); err != nil {
				return errors.Wrap(err, "execute returned action")
			}
		}

		return nil
	})
}
