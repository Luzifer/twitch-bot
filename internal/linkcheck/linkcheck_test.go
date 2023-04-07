package linkcheck

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestInfiniteRedirect(t *testing.T) {
	hdl := http.NewServeMux()
	hdl.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/test", http.StatusFound) })
	hdl.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/", http.StatusFound) })

	var (
		c  = New()
		ts = httptest.NewServer(hdl)
	)
	t.Cleanup(ts.Close)

	c.skipValidation = true

	msg := fmt.Sprintf("Here have a redirect loop: %s", ts.URL)

	// We expect /test to be the first repeat as the callstack will look like this:
	// ":12345", ":12345/test", ":12345/", ":12345/test" (which is the duplicate)
	assert.Equal(t, []string{fmt.Sprintf("%s/test", ts.URL)}, c.ScanForLinks(msg))
}

func TestMaxRedirects(t *testing.T) {
	hdl := mux.NewRouter()
	hdl.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/1", http.StatusFound) })
	hdl.HandleFunc("/{num}", func(w http.ResponseWriter, r *http.Request) {
		tn, _ := strconv.Atoi(mux.Vars(r)["num"])
		http.Redirect(w, r, fmt.Sprintf("/%d", tn+1), http.StatusFound)
	})

	var (
		c  = New()
		ts = httptest.NewServer(hdl)
	)
	t.Cleanup(ts.Close)

	c.skipValidation = true

	msg := fmt.Sprintf("Here have a redirect loop: %s", ts.URL)

	// We expect the call to `/N` to have N previous entries and therefore be the break-point
	assert.Equal(t, []string{fmt.Sprintf("%s/%d", ts.URL, maxRedirects)}, c.ScanForLinks(msg))
}

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
		// Case: obfuscation with "dot"
		{
			Message: "I'm live now on twitch dot tv/twitch",
			ExpectedLinks: []string{
				"https://www.twitch.tv/twitch",
			},
		},
		// Case: enhanced "dot" obfuscation
		{
			Message: "You can visit Archive(Dot) org in your browser",
			ExpectedLinks: []string{
				"http://Archive.org",
			},
		},
		// Case: false positive but not resolvable link
		{
			Message:       "game dot exe has stopped working",
			ExpectedLinks: nil,
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
