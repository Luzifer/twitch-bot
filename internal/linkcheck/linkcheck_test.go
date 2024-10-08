package linkcheck

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"testing"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestInfiniteRedirect(t *testing.T) {
	hdl := http.NewServeMux()
	hdl.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/test", http.StatusFound) })
	hdl.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/", http.StatusFound) })

	var (
		c  = New(withResolver(newResolver(1, withSkipVerify())))
		ts = httptest.NewServer(hdl)
	)
	t.Cleanup(ts.Close)

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
		c  = New(withResolver(newResolver(1, withSkipVerify())))
		ts = httptest.NewServer(hdl)
	)
	t.Cleanup(ts.Close)

	msg := fmt.Sprintf("Here have a redirect loop: %s", ts.URL)

	// We expect the call to `/N` to have N previous entries and therefore be the break-point
	assert.Equal(t, []string{fmt.Sprintf("%s/%d", ts.URL, maxRedirects)}, c.ScanForLinks(msg))
}

//nolint:funlen
func TestScanForLinks(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	for _, testCase := range []struct {
		Heuristic        bool
		Message          string
		ExpectedLinks    []string
		ExpectedContains bool
		TraceStack       bool
	}{
		// Case: full URL is present in the message
		{
			Heuristic: false,
			Message:   "https://example.com",
			ExpectedLinks: []string{
				"https://example.com",
			},
		},
		// Case: full bitly link is present in the message
		{
			Heuristic: false,
			Message:   "https://bit.ly/438obkJ",
			ExpectedLinks: []string{
				"https://example.com/",
			},
		},
		// Case: link is present just without the protocol
		{
			Heuristic: false,
			Message:   "Here, take a look at this: bit.ly/438obkJ",
			ExpectedLinks: []string{
				"https://example.com/",
			},
		},
		// Case: message with vk.cc shortener
		{
			Heuristic: false,
			Message:   "See more here: vk.cc/ckGZN2",
			ExpectedLinks: []string{
				"https://vk.com/club206261664",
			},
		},
		// Case: link is obfuscated using space
		{
			Heuristic: true,
			Message:   "Take a look on example. com",
			ExpectedLinks: []string{
				"http://example.com",
			},
			ExpectedContains: true,
		},
		// Case: link is obfuscated using space and braces
		{
			Heuristic: true,
			Message:   "Take a look on example. (com)",
			ExpectedLinks: []string{
				"http://example.com",
			},
			ExpectedContains: true,
		},
		// Case: multiple links in one message
		{
			Heuristic: false,
			Message:   "https://clips.twitch.tv/WrongEnchantingMinkFutureMan-EKlDjYkvDeurO9XT https://bit.ly/438obkJ",
			ExpectedLinks: []string{
				"https://clips.twitch.tv/WrongEnchantingMinkFutureMan-EKlDjYkvDeurO9XT",
				"https://example.com/",
			},
		},
		// Case: obfuscation with "dot"
		{
			Heuristic: true,
			Message:   "I'm live now on twitch dot tv/twitch",
			ExpectedLinks: []string{
				"https://www.twitch.tv/twitch",
			},
		},
		// Case: enhanced "dot" obfuscation
		{
			Heuristic: true,
			Message:   "You can visit Archive(Dot) org in your browser",
			ExpectedLinks: []string{
				"http://Archive.org",
			},
		},
		// Case: Youtube does weird stuff
		{
			Heuristic: false,
			Message:   "https://knut.in/youtube",
			ExpectedLinks: []string{
				"https://www.youtube.com/channel/UCjsRmaAQ0IHR2CNEBqfNOSQ",
			},
		},
		// Case: Instagram also does weird things
		{
			Heuristic: false,
			Message:   "https://bit.ly/3KHpJuy",
			ExpectedLinks: []string{
				"https://www.instagram.com/instagram/",
			},
		},
		// Case: Heuristic enabled with a German sentence accidentally
		// forming a valid link to a spanish site (btw.es) - known and
		// valid false-positive
		{
			Heuristic:     true,
			Message:       "Hey btw. es kann sein, dass",
			ExpectedLinks: []string{"https://trusted.evo-media.eu/btw.es"},
		},
		// Case: Multiple spaces in the link
		{
			Heuristic:        true,
			Message:          "Hey there, see my new project on exa mpl e. com! Get it fast now!",
			ExpectedLinks:    []string{"http://example.com"},
			ExpectedContains: true,
		},
		// Case: Dot in the end of the link with space
		{
			Heuristic:     true,
			Message:       "See example com. Nice testing stuff there!",
			ExpectedLinks: []string{"http://example.com"},
		},
		// Case: false positives
		{Heuristic: true, Message: "game dot exe has stopped working", ExpectedLinks: nil},
		{Heuristic: true, Message: "You are following since 12.12.2020 DogChamp", ExpectedLinks: nil},
		{Heuristic: false, Message: "Hey btw. es kann sein, dass", ExpectedLinks: nil},
	} {
		t.Run(fmt.Sprintf("h:%v lc:%d m:%s", testCase.Heuristic, len(testCase.ExpectedLinks), testCase.Message), func(t *testing.T) {
			var c *Checker
			if testCase.TraceStack {
				c = New(withResolver(newResolver(resolverPoolSize, withTesting(t))))
			} else {
				c = New()
			}

			var linksFound []string
			if testCase.Heuristic {
				linksFound = c.HeuristicScanForLinks(testCase.Message)
			} else {
				linksFound = c.ScanForLinks(testCase.Message)
			}
			sort.Strings(linksFound)

			if testCase.ExpectedContains {
				for _, expLnk := range testCase.ExpectedLinks {
					assert.Contains(t, linksFound, expLnk)
				}

				var extraLinks []string
				for _, link := range linksFound {
					if !str.StringInSlice(link, testCase.ExpectedLinks) {
						extraLinks = append(extraLinks, link)
					}
				}
				t.Logf("extra links found: %v", extraLinks)
			} else {
				assert.Equal(t, testCase.ExpectedLinks, linksFound)
			}
		})
	}
}
