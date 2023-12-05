package linkcheck

import (
	"context"
	"crypto/rand"
	_ "embed"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Luzifer/go_helpers/v2/str"
)

const (
	// DefaultCheckTimeout defines the default time the request to a site
	// may take to answer
	DefaultCheckTimeout = 10 * time.Second

	maxRedirects = 50
)

type (
	// Checker contains logic to detect and resolve links in a message
	Checker struct {
		checkTimeout time.Duration
		userAgents   []string

		skipValidation bool // Only for tests, not settable from the outside
	}
)

var (
	defaultUserAgents = []string{}
	linkTest          = regexp.MustCompile(`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`)
	numericHost       = regexp.MustCompile(`^(?:[0-9]+\.)*[0-9]+(?::[0-9]+)?$`)

	//go:embed user-agents.txt
	uaList string
)

func init() {
	defaultUserAgents = strings.Split(strings.TrimSpace(uaList), "\n")
}

// New creates a new Checker instance with default settings
func New() *Checker {
	return &Checker{
		checkTimeout: DefaultCheckTimeout,
		userAgents:   defaultUserAgents,
	}
}

// HeuristicScanForLinks takes a message and tries to find links
// within that message. Common methods like putting spaces into links
// are tried to circumvent.
func (c Checker) HeuristicScanForLinks(message string) []string {
	return c.scan(message,
		c.scanPlainNoObfuscate,
		c.scanDotObfuscation,
		c.scanObfuscateSpace,
		c.scanObfuscateSpecialCharsAndSpaces(regexp.MustCompile(`[^a-zA-Z0-9.:/\s_-]`), ""), // Leave dots intact and just join parts
		c.scanObfuscateSpecialCharsAndSpaces(regexp.MustCompile(`[^a-zA-Z0-9:/\s_-]`), "."), // Remove dots also and connect by them
	)
}

// ScanForLinks takes a message and tries to find links within that
// message. This only detects links without any means of obfuscation
// like putting spaces into the link.
func (c Checker) ScanForLinks(message string) (links []string) {
	return c.scan(message, c.scanPlainNoObfuscate)
}

// resolveFinal takes a link and looks up the final destination of
// that link after all redirects were followed
func (c Checker) resolveFinal(link string, cookieJar *cookiejar.Jar, callStack []string, userAgent string) string {
	if !linkTest.MatchString(link) && !c.skipValidation {
		return ""
	}

	if str.StringInSlice(link, callStack) || len(callStack) == maxRedirects {
		// We got ourselves a loop: Yay!
		return link
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: cookieJar,
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.checkTimeout)
	defer cancel()

	u, err := url.Parse(link)
	if err != nil {
		return ""
	}

	if u.Scheme == "" {
		// We have no scheme and the url is in the path, lets add the
		// scheme and re-parse the URL to avoid some confusion
		u.Scheme = "http"
		u, err = url.Parse(u.String())
		if err != nil {
			return ""
		}
	}

	if numericHost.MatchString(u.Host) && !c.skipValidation {
		// Host is fully numeric: We don't support scanning that
		return ""
	}

	// Sanitize host: Trailing dots are valid but not required
	u.Host = strings.TrimRight(u.Host, ".")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return ""
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 && resp.StatusCode < 400 {
		// We got a redirect
		tu, err := url.Parse(resp.Header.Get("location"))
		if err != nil {
			return ""
		}
		target := c.resolveReference(u, tu)
		return c.resolveFinal(target, cookieJar, append(callStack, link), userAgent)
	}

	// We got a response, it's no redirect, we count this as a success
	return u.String()
}

func (Checker) resolveReference(origin *url.URL, loc *url.URL) string {
	// Special Case: vkontakte used as shortener / obfuscation
	if loc.Path == "/away.php" && loc.Query().Has("to") {
		// VK is doing HTML / JS redirect magic so we take that from them
		// and execute the redirect directly here in code
		return loc.Query().Get("to")
	}

	if loc.Host == "consent.youtube.com" && loc.Query().Has("continue") {
		// Youtube links end up in consent page but we want the real
		// target so we use the continue parameter where we strip the
		// cbrd query parameters as that one causes an infinite loop.

		contTarget, err := url.Parse(loc.Query().Get("continue"))
		if err == nil {
			v := contTarget.Query()
			v.Del("cbrd")

			contTarget.RawQuery = v.Encode()
			return contTarget.String()
		}

		return loc.Query().Get("continue")
	}

	if loc.Host == "www.instagram.com" && loc.Query().Has("next") {
		// Instagram likes its login page, we on the other side don't
		// care about the sign-in or even the content. Therefore we
		// just take their redirect target and use that as the next
		// URL
		return loc.Query().Get("next")
	}

	// Default fallback behavior: Do a normal resolve
	return origin.ResolveReference(loc).String()
}

func (Checker) getJar() *cookiejar.Jar {
	jar, _ := cookiejar.New(nil)
	return jar
}

func (c Checker) scan(message string, scanFns ...func(string) []string) (links []string) {
	for _, scanner := range scanFns {
		if links = scanner(message); links != nil {
			return links
		}
	}

	return links
}

func (c Checker) scanDotObfuscation(message string) (links []string) {
	message = regexp.MustCompile(`(?i)\s*\(?dot\)?\s*`).ReplaceAllString(message, ".")
	return c.scanPlainNoObfuscate(message)
}

func (c Checker) scanObfuscateSpace(message string) (links []string) {
	// Spammers use spaces in their links to prevent link protection matches
	parts := regexp.MustCompile(`\s+`).Split(message, -1)
	return c.scanPartsConnected(parts, "")
}

func (c Checker) scanObfuscateSpecialCharsAndSpaces(set *regexp.Regexp, connector string) func(string) []string {
	return func(message string) (links []string) {
		// First clean URL from all characters not acceptable in Domains (plus some extra chars)
		message = set.ReplaceAllString(message, " ")
		parts := regexp.MustCompile(`\s+`).Split(message, -1)
		return c.scanPartsConnected(parts, connector)
	}
}

func (c Checker) scanPartsConnected(parts []string, connector string) (links []string) {
	for ptJoin := 2; ptJoin < len(parts); ptJoin++ {
		for i := 0; i <= len(parts)-ptJoin; i++ {
			if link := c.resolveFinal(strings.Join(parts[i:i+ptJoin], connector), c.getJar(), nil, c.userAgent()); link != "" && !str.StringInSlice(link, links) {
				links = append(links, link)
			}
		}
	}

	return links
}

func (c Checker) scanPlainNoObfuscate(message string) (links []string) {
	parts := regexp.MustCompile(`\s+`).Split(message, -1)

	for _, part := range parts {
		if link := c.resolveFinal(part, c.getJar(), nil, c.userAgent()); link != "" && !str.StringInSlice(link, links) {
			links = append(links, link)
		}
	}

	return links
}

func (c Checker) userAgent() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(c.userAgents))))
	return c.userAgents[n.Int64()]
}
