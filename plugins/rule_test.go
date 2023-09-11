package plugins

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
)

var (
	testLogger      = logrus.NewEntry(logrus.StandardLogger())
	testBadgeLevel0 = func(i int) *int { return &i }(0)
	testPtrBool     = func(b bool) *bool { return &b }
)

func TestAllowExecuteBadgeBlacklist(t *testing.T) {
	r := &Rule{DisableOn: []string{twitch.BadgeBroadcaster}}

	if r.allowExecuteBadgeBlacklist(testLogger, nil, nil, twitch.BadgeCollection{twitch.BadgeBroadcaster: testBadgeLevel0}, nil) {
		t.Error("Execution allowed on blacklisted badge")
	}

	if !r.allowExecuteBadgeBlacklist(testLogger, nil, nil, twitch.BadgeCollection{twitch.BadgeModerator: testBadgeLevel0}, nil) {
		t.Error("Execution denied without blacklisted badge")
	}
}

func TestAllowExecuteBadgeWhitelist(t *testing.T) {
	r := &Rule{EnableOn: []string{twitch.BadgeBroadcaster}}

	if r.allowExecuteBadgeWhitelist(testLogger, nil, nil, twitch.BadgeCollection{twitch.BadgeModerator: testBadgeLevel0}, nil) {
		t.Error("Execution allowed without whitelisted badge")
	}

	if !r.allowExecuteBadgeWhitelist(testLogger, nil, nil, twitch.BadgeCollection{twitch.BadgeBroadcaster: testBadgeLevel0}, nil) {
		t.Error("Execution denied with whitelisted badge")
	}
}

func TestAllowExecuteChannelWhitelist(t *testing.T) {
	r := &Rule{MatchChannels: []string{"#mychannel", "otherchannel"}}

	for m, exp := range map[string]bool{
		":amy!amy@foo.example.com PRIVMSG #mychannel :Testing":                                    true,
		":amy!amy@foo.example.com PRIVMSG #otherchannel :Testing":                                 true,
		":amy!amy@foo.example.com PRIVMSG #randomchannel :Testing":                                false,
		":amy!amy@foo.example.com JOIN #mychannel":                                                true,
		":tmi.twitch.tv CLEARCHAT #mychannel":                                                     true,
		":tmi.twitch.tv CLEARCHAT #mychannel :ronni":                                              true,
		":tmi.twitch.tv CLEARCHAT #dallas":                                                        false,
		"@msg-id=slow_off :tmi.twitch.tv NOTICE #mychannel :This room is no longer in slow mode.": true,
	} {
		if res := r.allowExecuteChannelWhitelist(testLogger, irc.MustParseMessage(m), nil, twitch.BadgeCollection{}, nil); res != exp {
			t.Errorf("Message %q yield unxpected result: exp=%v res=%v", m, exp, res)
		}
	}
}

func TestAllowExecuteDisable(t *testing.T) {
	for exp, r := range map[bool]*Rule{
		true:  {Disable: testPtrBool(false)},
		false: {Disable: testPtrBool(true)},
	} {
		if res := r.allowExecuteDisable(testLogger, nil, nil, twitch.BadgeCollection{}, nil); res != exp {
			t.Errorf("Disable status %v yield unexpected result: exp=%v res=%v", *r.Disable, exp, res)
		}
	}
}

func TestAllowExecuteDisableOnOffline(t *testing.T) {
	r := &Rule{DisableOnOffline: testPtrBool(true)}

	// Fake cache entries to prevent calling the real Twitch API
	r.twitchClient = twitch.New("", "", "", "")
	r.twitchClient.APICache().Set([]string{"hasLiveStream", "channel1"}, time.Minute, true)
	r.twitchClient.APICache().Set([]string{"hasLiveStream", "channel2"}, time.Minute, false)

	for ch, exp := range map[string]bool{
		"channel1": true,
		"channel2": false,
	} {
		if res := r.allowExecuteDisableOnOffline(testLogger, irc.MustParseMessage(fmt.Sprintf("PRIVMSG #%s :test", ch)), nil, twitch.BadgeCollection{}, nil); res != exp {
			t.Errorf("Channel %q yield an unexpected result: exp=%v res=%v", ch, exp, res)
		}
	}
}

func TestAllowExecuteChannelCooldown(t *testing.T) {
	r := &Rule{ChannelCooldown: func(i time.Duration) *time.Duration { return &i }(time.Minute), SkipCooldownFor: []string{twitch.BadgeBroadcaster}}
	c1 := irc.MustParseMessage(":amy!amy@foo.example.com PRIVMSG #mychannel :Testing")
	c2 := irc.MustParseMessage(":amy!amy@foo.example.com PRIVMSG #otherchannel :Testing")

	r.timerStore = newTestTimerStore()

	if !r.allowExecuteChannelCooldown(testLogger, c1, nil, twitch.BadgeCollection{}, nil) {
		t.Error("Initial call was not allowed")
	}

	// Add cooldown
	r.timerStore.AddCooldown(TimerTypeCooldown, c1.Params[0], r.MatcherID(), time.Now().Add(*r.ChannelCooldown))

	if r.allowExecuteChannelCooldown(testLogger, c1, nil, twitch.BadgeCollection{}, nil) {
		t.Error("Call after cooldown added was allowed")
	}

	if !r.allowExecuteChannelCooldown(testLogger, c1, nil, twitch.BadgeCollection{twitch.BadgeBroadcaster: testBadgeLevel0}, nil) {
		t.Error("Call in cooldown with skip badge was not allowed")
	}

	if !r.allowExecuteChannelCooldown(testLogger, c2, nil, twitch.BadgeCollection{twitch.BadgeBroadcaster: testBadgeLevel0}, nil) {
		t.Error("Call in cooldown with different channel was not allowed")
	}
}

func TestAllowExecuteDisableOnPermit(t *testing.T) {
	r := &Rule{DisableOnPermit: testPtrBool(true)}
	r.timerStore = newTestTimerStore()

	m := irc.MustParseMessage(":amy!amy@foo.example.com PRIVMSG #mychannel :Testing")
	if !r.allowExecuteDisableOnPermit(testLogger, m, nil, twitch.BadgeCollection{}, nil) {
		t.Error("Execution was not allowed without permit")
	}

	r.timerStore.AddPermit(m.Params[0], m.User)
	if r.allowExecuteDisableOnPermit(testLogger, m, nil, twitch.BadgeCollection{}, nil) {
		t.Error("Execution was allowed with permit")
	}
}

func TestAllowExecuteDisableOnTemplate(t *testing.T) {
	r := &Rule{DisableOnTemplate: func(s string) *string { return &s }(`{{ ne .username "amy" }}`)}

	for msg, exp := range map[string]bool{
		"false": true,
		"true":  false,
	} {
		// We don't test the message formatter here but only the disable functionality
		// so we fake the result of the evaluation
		r.msgFormatter = func(tplString string, m *irc.Message, r *Rule, fields *FieldCollection) (string, error) {
			return msg, nil
		}

		if res := r.allowExecuteDisableOnTemplate(testLogger, irc.MustParseMessage(msg), nil, twitch.BadgeCollection{}, nil); exp != res {
			t.Errorf("Message %q yield unexpected result: exp=%v res=%v", msg, exp, res)
		}
	}
}

func TestAllowExecuteEventWhitelist(t *testing.T) {
	r := &Rule{MatchEvent: func(s string) *string { return &s }("test")}

	for evt, exp := range map[string]bool{
		"foobar": false,
		"test":   true,
	} {
		if res := r.allowExecuteEventMatch(testLogger, nil, &evt, twitch.BadgeCollection{}, nil); exp != res {
			t.Errorf("Event %q yield unexpected result: exp=%v res=%v", evt, exp, res)
		}
	}
}

func TestAllowExecuteMessageMatcherBlacklist(t *testing.T) {
	r := &Rule{DisableOnMatchMessages: []string{`^!disable`}}

	for msg, exp := range map[string]bool{
		"PRIVMSG #test :Random message":    true,
		"PRIVMSG #test :!disable this one": false,
	} {
		if res := r.allowExecuteMessageMatcherBlacklist(testLogger, irc.MustParseMessage(msg), nil, twitch.BadgeCollection{}, nil); exp != res {
			t.Errorf("Message %q yield unexpected result: exp=%v res=%v", msg, exp, res)
		}
	}
}

func TestAllowExecuteMessageMatcherWhitelist(t *testing.T) {
	r := &Rule{MatchMessage: func(s string) *string { return &s }(`^!test`)}

	for msg, exp := range map[string]bool{
		"PRIVMSG #test :Random message": false,
		"PRIVMSG #test :!test this one": true,
	} {
		if res := r.allowExecuteMessageMatcherWhitelist(testLogger, irc.MustParseMessage(msg), nil, twitch.BadgeCollection{}, nil); exp != res {
			t.Errorf("Message %q yield unexpected result: exp=%v res=%v", msg, exp, res)
		}
	}
}

func TestAllowExecuteRuleCooldown(t *testing.T) {
	r := &Rule{Cooldown: func(i time.Duration) *time.Duration { return &i }(time.Minute), SkipCooldownFor: []string{twitch.BadgeBroadcaster}}
	r.timerStore = newTestTimerStore()

	if !r.allowExecuteRuleCooldown(testLogger, nil, nil, twitch.BadgeCollection{}, nil) {
		t.Error("Initial call was not allowed")
	}

	// Add cooldown
	r.timerStore.AddCooldown(TimerTypeCooldown, "", r.MatcherID(), time.Now().Add(*r.Cooldown))

	if r.allowExecuteRuleCooldown(testLogger, nil, nil, twitch.BadgeCollection{}, nil) {
		t.Error("Call after cooldown added was allowed")
	}

	if !r.allowExecuteRuleCooldown(testLogger, nil, nil, twitch.BadgeCollection{twitch.BadgeBroadcaster: testBadgeLevel0}, nil) {
		t.Error("Call in cooldown with skip badge was not allowed")
	}
}

func TestAllowExecuteUserCooldown(t *testing.T) {
	r := &Rule{UserCooldown: func(i time.Duration) *time.Duration { return &i }(time.Minute), SkipCooldownFor: []string{twitch.BadgeBroadcaster}}
	c1 := irc.MustParseMessage(":ben!ben@foo.example.com PRIVMSG #mychannel :Testing")
	c2 := irc.MustParseMessage(":amy!amy@foo.example.com PRIVMSG #mychannel :Testing")

	r.timerStore = newTestTimerStore()

	if !r.allowExecuteUserCooldown(testLogger, c1, nil, twitch.BadgeCollection{}, nil) {
		t.Error("Initial call was not allowed")
	}

	// Add cooldown
	r.timerStore.AddCooldown(TimerTypeCooldown, c1.User, r.MatcherID(), time.Now().Add(*r.UserCooldown))

	if r.allowExecuteUserCooldown(testLogger, c1, nil, twitch.BadgeCollection{}, nil) {
		t.Error("Call after cooldown added was allowed")
	}

	if !r.allowExecuteUserCooldown(testLogger, c1, nil, twitch.BadgeCollection{twitch.BadgeBroadcaster: testBadgeLevel0}, nil) {
		t.Error("Call in cooldown with skip badge was not allowed")
	}

	if !r.allowExecuteUserCooldown(testLogger, c2, nil, twitch.BadgeCollection{twitch.BadgeBroadcaster: testBadgeLevel0}, nil) {
		t.Error("Call in cooldown with different user was not allowed")
	}
}

func TestAllowExecuteUserWhitelist(t *testing.T) {
	r := &Rule{MatchUsers: []string{"amy"}}

	for msg, exp := range map[string]bool{
		":amy!amy@foo.example.com PRIVMSG #mychannel :Testing": true,
		":bob!bob@foo.example.com PRIVMSG #mychannel :Testing": false,
	} {
		if res := r.allowExecuteUserWhitelist(testLogger, irc.MustParseMessage(msg), nil, twitch.BadgeCollection{}, nil); exp != res {
			t.Errorf("Message %q yield unexpected result: exp=%v res=%v", msg, exp, res)
		}
	}
}
