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
	dropSet           = regexp.MustCompile(`[^a-zA-Z0-9.:/\s_-]`)
	linkTest          = regexp.MustCompile(`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`)

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

// ScanForLinks takes a message and tries to find links within that
// message. Common methods like putting spaces into links are tried
// to circumvent.
func (c Checker) ScanForLinks(message string) (links []string) {
	for _, scanner := range []func(string) []string{
		c.scanPlainNoObfuscate,
		c.scanObfuscateSpace,
		c.scanObfuscateSpecialCharsAndSpaces,
		c.scanDotObfuscation,
	} {
		if links = scanner(message); links != nil {
			return links
		}
	}

	return links
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
		u.Scheme = "http"
	}

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

	// Default fallback behavior: Do a normal resolve
	return origin.ResolveReference(loc).String()
}

func (Checker) getJar() *cookiejar.Jar {
	jar, _ := cookiejar.New(nil)
	return jar
}

func (c Checker) scanDotObfuscation(message string) (links []string) {
	message = regexp.MustCompile(`(?i)\s*\(?dot\)?\s*`).ReplaceAllString(message, ".")
	return c.scanPlainNoObfuscate(message)
}

func (c Checker) scanObfuscateSpace(message string) (links []string) {
	// Spammers use spaces in their links to prevent link protection matches
	parts := regexp.MustCompile(`\s+`).Split(message, -1)

	for i := 0; i < len(parts)-1; i++ {
		if link := c.resolveFinal(strings.Join(parts[i:i+2], ""), c.getJar(), nil, c.userAgent()); link != "" {
			links = append(links, link)
		}
	}

	return links
}

func (c Checker) scanObfuscateSpecialCharsAndSpaces(message string) (links []string) {
	// First clean URL from all characters not acceptable in Domains (plus some extra chars)
	message = dropSet.ReplaceAllString(message, "")
	return c.scanObfuscateSpace(message)
}

func (c Checker) scanPlainNoObfuscate(message string) (links []string) {
	parts := regexp.MustCompile(`\s+`).Split(message, -1)

	for _, part := range parts {
		if link := c.resolveFinal(part, c.getJar(), nil, c.userAgent()); link != "" {
			links = append(links, link)
		}
	}

	return links
}

func (c Checker) userAgent() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(c.userAgents))))
	return c.userAgents[n.Int64()]
}
