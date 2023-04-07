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
		// Case: message with vk.cc shortener
		{
			Message: "See more here: vk.cc/ckGZN2",
			ExpectedLinks: []string{
				"https://vk.com/club206261664",
			},
		},
		// Case: link is obfuscated using space
		{
			Message: "Take a look at example. com",
			ExpectedLinks: []string{
				"http://example.com",
			},
		},
		// Case: link is obfuscated using space and braces
		{
			Message: "Take a look at example. (com)",
			ExpectedLinks: []string{
				"http://example.com",
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
