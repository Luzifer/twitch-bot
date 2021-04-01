package main

import (
	"testing"

	"github.com/sirupsen/logrus"
)

var (
	testLogger      = logrus.NewEntry(logrus.StandardLogger())
	testBadgeLevel0 = func(i int) *int { return &i }(0)
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

func TestAllowExecuteChannelWhitelist(t *testing.T)        { t.Fatal("Not implemented") }
func TestAllowExecuteCooldown(t *testing.T)                { t.Fatal("Not implemented") }
func TestAllowExecuteDisableOnOffline(t *testing.T)        { t.Fatal("Not implemented") }
func TestAllowExecuteDisableOnPermit(t *testing.T)         { t.Fatal("Not implemented") }
func TestAllowExecuteEventWhitelist(t *testing.T)          { t.Fatal("Not implemented") }
func TestAllowExecuteMessageMatcherBlacklist(t *testing.T) { t.Fatal("Not implemented") }
func TestAllowExecuteMessageMatcherWhitelist(t *testing.T) { t.Fatal("Not implemented") }
func TestAllowExecuteUserWhitelist(t *testing.T)           { t.Fatal("Not implemented") }
