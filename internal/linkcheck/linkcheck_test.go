package linkcheck

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"sort"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/Luzifer/twitch-bot/v3/internal/linkcheck/fixture"
)

type (
	scanForLinksCase struct {
		Heuristic        bool
		Message          string
		ExpectedLinks    []string
		ExpectedContains bool
		NoLive           bool
		TraceStack       bool
	}
)

var scanForLinksCases = []scanForLinksCase{
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
			"https://www.instagram.com/instagram",
		},
	},
	// Case: Heuristic enabled with a German sentence accidentally
	// forming a valid link to a spanish site (btw.es) - known and
	// valid false-positive
	{
		Heuristic:     true,
		Message:       "Hey btw. es kann sein, dass",
		ExpectedLinks: []string{"https://trusted.domainseller.site/btw.es"},
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
	// Broken cloudflare configuration but valid match for a link
	{
		Heuristic:     false,
		Message:       "http://http521.invalid/",
		ExpectedLinks: []string{"http://http521.invalid/"},
		NoLive:        true,
	},
	{
		Heuristic:     false,
		Message:       "http://metaredirect.invalid/",
		ExpectedLinks: []string{"https://example.com/"},
		NoLive:        true,
	},
	{
		Message:       "https://vk.cc/hfOwt",
		ExpectedLinks: []string{"http://yandex.ru/"},
	},
	{
		Message:       "https://cookieguard.invalid/",
		ExpectedLinks: []string{"https://example.com/"},
		NoLive:        true,
	},
	// Case: false positives
	{Heuristic: true, Message: "game dot exe has stopped working", ExpectedLinks: nil},
	{Heuristic: true, Message: "You are following since 12.12.2020 DogChamp", ExpectedLinks: nil},
	{Heuristic: false, Message: "Hey btw. es kann sein, dass", ExpectedLinks: nil},
}

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

//nolint:gocognit // handles different case types
func TestScanForLinks(t *testing.T) {
	isLiveCheckUsingInternet := slices.Contains([]string{"1", "true"}, os.Getenv("LINKCHECK_LIVE_TESTS"))

	if testing.Short() && isLiveCheckUsingInternet {
		t.SkipNow()
	}

	for _, testCase := range scanForLinksCases {
		t.Run(fmt.Sprintf("h:%v lc:%d m:%s", testCase.Heuristic, len(testCase.ExpectedLinks), testCase.Message), func(t *testing.T) {
			var resolverOpts []func(*resolver)

			if testCase.TraceStack {
				resolverOpts = append(resolverOpts, withTesting(t))
			}

			if !isLiveCheckUsingInternet {
				resolverOpts = append(resolverOpts, withTransport(fixture.New()))
			} else if testCase.NoLive {
				t.Skip("disabled in live-check")
			}

			c := New(withResolver(newResolver(resolverPoolSize, resolverOpts...)))

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
					if !slices.Contains(testCase.ExpectedLinks, link) {
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
