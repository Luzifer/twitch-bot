package main

import (
	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func() Actor { return &ActorSetVariable{} })
}

type ActorSetVariable struct {
	Variable string `json:"variable" yaml:"variable"`
	Clear    bool   `json:"clear" yaml:"clear"`
	Set      string `json:"set" yaml:"set"`
}

func (a ActorSetVariable) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule) (preventCooldown bool, err error) {
	if a.Variable == "" {
		return false, nil
	}

	varName, err := formatMessage(a.Variable, m, r, nil)
	if err != nil {
		return false, errors.Wrap(err, "preparing variable name")
	}

	if a.Clear {
		return false, errors.Wrap(
			store.RemoveVariable(varName),
			"removing variable",
		)
	}

	value, err := formatMessage(a.Set, m, r, nil)
	if err != nil {
		return false, errors.Wrap(err, "preparing value")
	}

	return false, errors.Wrap(
		store.SetVariable(varName, value),
		"setting variable",
	)
}

func (a ActorSetVariable) IsAsync() bool { return false }
func (a ActorSetVariable) Name() string  { return "setvariable" }
