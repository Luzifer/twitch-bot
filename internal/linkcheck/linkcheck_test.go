package linkcheck

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanForLinks(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	c := New()

	for _, testCase := range []struct {
		Message       string
		ExpectedLinks []string
	}{
		// Case: full URL is present in the message
		{
			Message: "https://example.com",
			ExpectedLinks: []string{
				"https://example.com",
			},
		},
		// Case: full bitly link is present in the message
		{
			Message: "https://bit.ly/438obkJ",
			ExpectedLinks: []string{
				"https://example.com/",
			},
		},
		// Case: link is present just without the protocol
		{
			Message: "Here, take a look at this: bit.ly/438obkJ",
			ExpectedLinks: []string{
				"https://example.com/",
			},
		},
		// Case: typical spam message with vk.cc shortener
		{
			Message: "Wanna become famous?  Upgrade your channel on vk.cc/cmskar",
			ExpectedLinks: []string{
				"https://goodness.selly.store",
			},
		},
		// Case: link is obfuscated using space
		{
			Message: "Wow! Upgrade your channel on yourfollowz. com",
			ExpectedLinks: []string{
				"http://yourfollowz.com",
			},
		},
		// Case: link is obfuscated using space and braces
		{
			Message: "Wow! Upgrade your channel on yourfollowz. (com)",
			ExpectedLinks: []string{
				"http://yourfollowz.com",
			},
		},
		// Case: multiple links in one message
		{
			Message: "https://clips.twitch.tv/WrongEnchantingMinkFutureMan-EKlDjYkvDeurO9XT https://bit.ly/438obkJ",
			ExpectedLinks: []string{
				"https://clips.twitch.tv/WrongEnchantingMinkFutureMan-EKlDjYkvDeurO9XT",
				"https://example.com/",
			},
		},
	} {
		linksFound := c.ScanForLinks(testCase.Message)
		sort.Strings(linksFound)

		assert.Equal(t, testCase.ExpectedLinks, linksFound, "links from message %q", testCase.Message)
	}
}

func TestUserAgentListNotEmpty(t *testing.T) {
	if len(defaultUserAgents) == 0 {
		t.Fatal("found empty user-agent list")
	}
}

func TestUserAgentRandomizer(t *testing.T) {
	var (
		c   = New()
		uas = map[string]int{}
	)

	for i := 0; i < 10; i++ {
		uas[c.userAgent()]++
	}

	for _, c := range uas {
		assert.Less(t, c, 10)
	}

	assert.Equal(t, 0, uas[""]) // there should be no empty UA
}
