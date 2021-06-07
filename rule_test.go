package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-irc/irc"
	"github.com/sirupsen/logrus"
)

var (
	testLogger      = logrus.NewEntry(logrus.StandardLogger())
	testBadgeLevel0 = func(i int) *int { return &i }(0)
	testPtrBool     = func(b bool) *bool { return &b }
)

func TestAllowExecuteBadgeBlacklist(t *testing.T) {
	r := &rule{DisableOn: []string{badgeBroadcaster}}

	if r.allowExecuteBadgeBlacklist(testLogger, nil, nil, badgeCollection{badgeBroadcaster: testBadgeLevel0}) {
		t.Error("Execution allowed on blacklisted badge")
	}

	if !r.allowExecuteBadgeBlacklist(testLogger, nil, nil, badgeCollection{badgeModerator: testBadgeLevel0}) {
		t.Error("Execution denied without blacklisted badge")
	}
}

func TestAllowExecuteBadgeWhitelist(t *testing.T) {
	r := &rule{EnableOn: []string{badgeBroadcaster}}

	if r.allowExecuteBadgeWhitelist(testLogger, nil, nil, badgeCollection{badgeModerator: testBadgeLevel0}) {
		t.Error("Execution allowed without whitelisted badge")
	}

	if !r.allowExecuteBadgeWhitelist(testLogger, nil, nil, badgeCollection{badgeBroadcaster: testBadgeLevel0}) {
		t.Error("Execution denied with whitelisted badge")
	}
}

func TestAllowExecuteChannelWhitelist(t *testing.T) {
	r := &rule{MatchChannels: []string{"#mychannel", "otherchannel"}}

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
		if res := r.allowExecuteChannelWhitelist(testLogger, irc.MustParseMessage(m), nil, badgeCollection{}); res != exp {
			t.Errorf("Message %q yield unxpected result: exp=%v res=%v", m, exp, res)
		}
	}
}

func TestAllowExecuteDisable(t *testing.T) {
	for exp, r := range map[bool]*rule{
		true:  {Disable: testPtrBool(false)},
		false: {Disable: testPtrBool(true)},
	} {
		if res := r.allowExecuteDisable(testLogger, nil, nil, badgeCollection{}); res != exp {
			t.Errorf("Disable status %v yield unexpected result: exp=%v res=%v", *r.Disable, exp, res)
		}
	}
}

func TestAllowExecuteDisableOnOffline(t *testing.T) {
	r := &rule{DisableOnOffline: testPtrBool(true)}

	// Fake cache entries to prevent calling the real Twitch API
	twitch.apiCache.Set([]string{"hasLiveStream", "channel1"}, time.Minute, true)
	twitch.apiCache.Set([]string{"hasLiveStream", "channel2"}, time.Minute, false)

	for ch, exp := range map[string]bool{
		"channel1": true,
		"channel2": false,
	} {
		if res := r.allowExecuteDisableOnOffline(testLogger, irc.MustParseMessage(fmt.Sprintf("PRIVMSG #%s :test", ch)), nil, badgeCollection{}); res != exp {
			t.Errorf("Channel %q yield an unexpected result: exp=%v res=%v", ch, exp, res)
		}
	}
}

func TestAllowExecuteChannelCooldown(t *testing.T) {
	r := &rule{ChannelCooldown: func(i time.Duration) *time.Duration { return &i }(time.Minute), SkipCooldownFor: []string{badgeBroadcaster}}
	c1 := irc.MustParseMessage(":amy!amy@foo.example.com PRIVMSG #mychannel :Testing")
	c2 := irc.MustParseMessage(":amy!amy@foo.example.com PRIVMSG #otherchannel :Testing")

	if !r.allowExecuteChannelCooldown(testLogger, c1, nil, badgeCollection{}) {
		t.Error("Initial call was not allowed")
	}

	// Add cooldown
	timerStore.AddCooldown(timerTypeCooldown, c1.Params[0], r.MatcherID())

	if r.allowExecuteChannelCooldown(testLogger, c1, nil, badgeCollection{}) {
		t.Error("Call after cooldown added was allowed")
	}

	if !r.allowExecuteChannelCooldown(testLogger, c1, nil, badgeCollection{badgeBroadcaster: testBadgeLevel0}) {
		t.Error("Call in cooldown with skip badge was not allowed")
	}

	if !r.allowExecuteChannelCooldown(testLogger, c2, nil, badgeCollection{badgeBroadcaster: testBadgeLevel0}) {
		t.Error("Call in cooldown with different channel was not allowed")
	}
}

func TestAllowExecuteDisableOnPermit(t *testing.T) {
	r := &rule{DisableOnPermit: testPtrBool(true)}

	// Permit is using global configuration, so we must fake that one
	config = &configFile{PermitTimeout: time.Minute}
	defer func() { config = nil }()

	m := irc.MustParseMessage(":amy!amy@foo.example.com PRIVMSG #mychannel :Testing")
	if !r.allowExecuteDisableOnPermit(testLogger, m, nil, badgeCollection{}) {
		t.Error("Execution was not allowed without permit")
	}

	timerStore.AddPermit(m.Params[0], m.User)
	if r.allowExecuteDisableOnPermit(testLogger, m, nil, badgeCollection{}) {
		t.Error("Execution was allowed with permit")
	}
}

func TestAllowExecuteDisableOnTemplate(t *testing.T) {
	r := &rule{DisableOnTemplate: func(s string) *string { return &s }(`{{ ne .username "amy" }}`)}

	for msg, exp := range map[string]bool{
		":amy!amy@foo.example.com PRIVMSG #mychannel :Testing": true,
		":bob!bob@foo.example.com PRIVMSG #mychannel :Testing": false,
	} {
		if res := r.allowExecuteDisableOnTemplate(testLogger, irc.MustParseMessage(msg), nil, badgeCollection{}); exp != res {
			t.Errorf("Message %q yield unexpected result: exp=%v res=%v", msg, exp, res)
		}
	}
}

func TestAllowExecuteEventWhitelist(t *testing.T) {
	r := &rule{MatchEvent: func(s string) *string { return &s }("test")}

	for evt, exp := range map[string]bool{
		"foobar": false,
		"test":   true,
	} {
		if res := r.allowExecuteEventWhitelist(testLogger, nil, &evt, badgeCollection{}); exp != res {
			t.Errorf("Event %q yield unexpected result: exp=%v res=%v", evt, exp, res)
		}
	}
}

func TestAllowExecuteMessageMatcherBlacklist(t *testing.T) {
	r := &rule{DisableOnMatchMessages: []string{`^!disable`}}

	for msg, exp := range map[string]bool{
		"PRIVMSG #test :Random message":    true,
		"PRIVMSG #test :!disable this one": false,
	} {
		if res := r.allowExecuteMessageMatcherBlacklist(testLogger, irc.MustParseMessage(msg), nil, badgeCollection{}); exp != res {
			t.Errorf("Message %q yield unexpected result: exp=%v res=%v", msg, exp, res)
		}
	}
}

func TestAllowExecuteMessageMatcherWhitelist(t *testing.T) {
	r := &rule{MatchMessage: func(s string) *string { return &s }(`^!test`)}

	for msg, exp := range map[string]bool{
		"PRIVMSG #test :Random message": false,
		"PRIVMSG #test :!test this one": true,
	} {
		if res := r.allowExecuteMessageMatcherWhitelist(testLogger, irc.MustParseMessage(msg), nil, badgeCollection{}); exp != res {
			t.Errorf("Message %q yield unexpected result: exp=%v res=%v", msg, exp, res)
		}
	}
}

func TestAllowExecuteRuleCooldown(t *testing.T) {
	r := &rule{Cooldown: func(i time.Duration) *time.Duration { return &i }(time.Minute), SkipCooldownFor: []string{badgeBroadcaster}}

	if !r.allowExecuteRuleCooldown(testLogger, nil, nil, badgeCollection{}) {
		t.Error("Initial call was not allowed")
	}

	// Add cooldown
	timerStore.AddCooldown(timerTypeCooldown, "", r.MatcherID())

	if r.allowExecuteRuleCooldown(testLogger, nil, nil, badgeCollection{}) {
		t.Error("Call after cooldown added was allowed")
	}

	if !r.allowExecuteRuleCooldown(testLogger, nil, nil, badgeCollection{badgeBroadcaster: testBadgeLevel0}) {
		t.Error("Call in cooldown with skip badge was not allowed")
	}
}

func TestAllowExecuteUserCooldown(t *testing.T) {
	r := &rule{UserCooldown: func(i time.Duration) *time.Duration { return &i }(time.Minute), SkipCooldownFor: []string{badgeBroadcaster}}
	c1 := irc.MustParseMessage(":ben!ben@foo.example.com PRIVMSG #mychannel :Testing")
	c2 := irc.MustParseMessage(":amy!amy@foo.example.com PRIVMSG #mychannel :Testing")

	if !r.allowExecuteUserCooldown(testLogger, c1, nil, badgeCollection{}) {
		t.Error("Initial call was not allowed")
	}

	// Add cooldown
	timerStore.AddCooldown(timerTypeCooldown, c1.User, r.MatcherID())

	if r.allowExecuteUserCooldown(testLogger, c1, nil, badgeCollection{}) {
		t.Error("Call after cooldown added was allowed")
	}

	if !r.allowExecuteUserCooldown(testLogger, c1, nil, badgeCollection{badgeBroadcaster: testBadgeLevel0}) {
		t.Error("Call in cooldown with skip badge was not allowed")
	}

	if !r.allowExecuteUserCooldown(testLogger, c2, nil, badgeCollection{badgeBroadcaster: testBadgeLevel0}) {
		t.Error("Call in cooldown with different user was not allowed")
	}
}

func TestAllowExecuteUserWhitelist(t *testing.T) {
	r := &rule{MatchUsers: []string{"amy"}}

	for msg, exp := range map[string]bool{
		":amy!amy@foo.example.com PRIVMSG #mychannel :Testing": true,
		":bob!bob@foo.example.com PRIVMSG #mychannel :Testing": false,
	} {
		if res := r.allowExecuteUserWhitelist(testLogger, irc.MustParseMessage(msg), nil, badgeCollection{}); exp != res {
			t.Errorf("Message %q yield unexpected result: exp=%v res=%v", msg, exp, res)
		}
	}
}
