package main

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/stretchr/testify/require"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

type testExecutionActor struct {
	exec func(*irc.Message) (bool, error)
	name string
}

func (a testExecutionActor) Execute(_ *irc.Client, m *irc.Message, _ *plugins.Rule, _ *fieldcollection.FieldCollection, _ *fieldcollection.FieldCollection) (bool, error) {
	return a.exec(m)
}

func (testExecutionActor) IsAsync() bool { return false }

func (a testExecutionActor) Name() string { return a.name }

func (testExecutionActor) Validate(plugins.TemplateValidatorFunc, *fieldcollection.FieldCollection) error {
	return nil
}

func registerTestExecutionActor(t *testing.T, exec func(*irc.Message) (bool, error)) string {
	t.Helper()

	name := fmt.Sprintf("test-execution-%s-%d", t.Name(), time.Now().UnixNano())
	registerAction(name, func() plugins.Actor { return testExecutionActor{name: name, exec: exec} })

	return name
}

func testExecutionRule(actionName, uuid string) *plugins.Rule {
	return &plugins.Rule{
		UUID: uuid,
		Actions: []*plugins.RuleAction{
			{Type: actionName, Attributes: fieldcollection.NewFieldCollection()},
		},
	}
}

func TestHandleMessageRuleExecutionAppliesRuleCooldownOnce(t *testing.T) {
	var executed atomic.Int32

	actionName := registerTestExecutionActor(t, func(*irc.Message) (bool, error) {
		executed.Add(1)
		time.Sleep(50 * time.Millisecond)
		return false, nil
	})

	rule := testExecutionRule(actionName, "rule-cooldown-once")
	rule.Cooldown = func(d time.Duration) *time.Duration { return &d }(time.Minute)
	msg := irc.MustParseMessage(":amy!amy@foo.example.com PRIVMSG #mychannel :!test")

	var wg sync.WaitGroup
	wg.Add(2)

	for range 2 {
		go func() {
			defer wg.Done()
			handleMessageRuleExecution(nil, msg, rule, nil)
		}()
	}

	wg.Wait()

	require.Equal(t, int32(1), executed.Load())

	inCooldown, err := timerService.InCooldown(plugins.TimerTypeCooldown, "", rule.MatcherID())
	require.NoError(t, err)
	require.True(t, inCooldown)
}

func TestHandleMessageRuleExecutionAppliesChannelCooldownOncePerChannel(t *testing.T) {
	var executed atomic.Int32

	actionName := registerTestExecutionActor(t, func(*irc.Message) (bool, error) {
		executed.Add(1)
		time.Sleep(50 * time.Millisecond)
		return false, nil
	})

	rule := testExecutionRule(actionName, "channel-cooldown-once")
	rule.ChannelCooldown = func(d time.Duration) *time.Duration { return &d }(time.Minute)
	msg := irc.MustParseMessage(":amy!amy@foo.example.com PRIVMSG #mychannel :!test")

	var wg sync.WaitGroup
	wg.Add(2)

	for range 2 {
		go func() {
			defer wg.Done()
			handleMessageRuleExecution(nil, msg, rule, nil)
		}()
	}

	wg.Wait()

	require.Equal(t, int32(1), executed.Load())

	inCooldown, err := timerService.InCooldown(plugins.TimerTypeCooldown, "#mychannel", rule.MatcherID())
	require.NoError(t, err)
	require.True(t, inCooldown)
}

func TestHandleMessageRuleExecutionAllowsChannelCooldownAcrossChannels(t *testing.T) {
	var executed atomic.Int32

	actionName := registerTestExecutionActor(t, func(*irc.Message) (bool, error) {
		executed.Add(1)
		time.Sleep(50 * time.Millisecond)
		return false, nil
	})

	rule := testExecutionRule(actionName, "channel-cooldown-different-channels")
	rule.ChannelCooldown = func(d time.Duration) *time.Duration { return &d }(time.Minute)

	msgA := irc.MustParseMessage(":amy!amy@foo.example.com PRIVMSG #mychannel :!test")
	msgB := irc.MustParseMessage(":amy!amy@foo.example.com PRIVMSG #otherchannel :!test")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		handleMessageRuleExecution(nil, msgA, rule, nil)
	}()

	go func() {
		defer wg.Done()
		handleMessageRuleExecution(nil, msgB, rule, nil)
	}()

	wg.Wait()

	require.Equal(t, int32(2), executed.Load())
}

func TestHandleMessageRuleExecutionSkipsCooldownWhenPrevented(t *testing.T) {
	actionName := registerTestExecutionActor(t, func(*irc.Message) (bool, error) {
		return true, nil
	})

	rule := testExecutionRule(actionName, "prevent-cooldown")
	rule.Cooldown = func(d time.Duration) *time.Duration { return &d }(time.Minute)
	msg := irc.MustParseMessage(":amy!amy@foo.example.com PRIVMSG #mychannel :!test")

	handleMessageRuleExecution(nil, msg, rule, nil)

	inCooldown, err := timerService.InCooldown(plugins.TimerTypeCooldown, "", rule.MatcherID())
	require.NoError(t, err)
	require.False(t, inCooldown)
}

func TestHandleMessageRuleExecutionSkipsCooldownOnActionError(t *testing.T) {
	actionName := registerTestExecutionActor(t, func(*irc.Message) (bool, error) {
		return false, errors.New("boom")
	})

	rule := testExecutionRule(actionName, "error-no-cooldown")
	rule.Cooldown = func(d time.Duration) *time.Duration { return &d }(time.Minute)
	msg := irc.MustParseMessage(":amy!amy@foo.example.com PRIVMSG #mychannel :!test")

	handleMessageRuleExecution(nil, msg, rule, nil)

	inCooldown, err := timerService.InCooldown(plugins.TimerTypeCooldown, "", rule.MatcherID())
	require.NoError(t, err)
	require.False(t, inCooldown)
}
