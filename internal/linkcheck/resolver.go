package linkcheck

import (
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// DefaultCheckTimeout defines the default time the request to a site
	// may take to answer
	DefaultCheckTimeout = 10 * time.Second

	maxRedirects     = 50
	resolverPoolSize = 25
)

type (
	resolver struct {
		resolverC      chan resolverQueueEntry
		skipValidation bool

		t *testing.T
	}

	resolverQueueEntry struct {
		Link      string
		Callback  func(string)
		WaitGroup *sync.WaitGroup
	}
)

var (
	linkTest    = regexp.MustCompile(`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`)
	numericHost = regexp.MustCompile(`^(?:[0-9]+\.)*[0-9]+(?::[0-9]+)?$`)

	defaultResolver = newResolver(resolverPoolSize)
)

func newResolver(poolSize int, opts ...func(*resolver)) *resolver {
	r := &resolver{
		resolverC: make(chan resolverQueueEntry),
	}

	for _, o := range opts {
		o(r)
	}

	for i := 0; i < poolSize; i++ {
		go r.runResolver()
	}

	return r
}

func withSkipVerify() func(*resolver) {
	return func(r *resolver) { r.skipValidation = true }
}

func withTesting(t *testing.T) func(*resolver) {
	return func(r *resolver) { r.t = t }
}

func (r resolver) Resolve(qe resolverQueueEntry) {
	qe.WaitGroup.Add(1)
	r.resolverC <- qe
}

func (resolver) getJar() *cookiejar.Jar {
	jar, _ := cookiejar.New(nil)
	return jar
}

// resolveFinal takes a link and looks up the final destination of
// that link after all redirects were followed
//
//nolint:funlen,gocyclo
func (r resolver) resolveFinal(link string, cookieJar *cookiejar.Jar, callStack *stack) string {
	if !linkTest.MatchString(link) && !r.skipValidation {
		return ""
	}

	if callStack.Count(link) > 2 || callStack.Height() == maxRedirects {
		// We got ourselves a loop: Yay!
		return link
	}

	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: cookieJar,
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultCheckTimeout)
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

	if numericHost.MatchString(u.Host) && !r.skipValidation {
		// Host is fully numeric: We don't support scanning that
		return ""
	}

	// Sanitize host: Trailing dots are valid but not required
	u.Host = strings.TrimRight(u.Host, ".")

	if r.t != nil {
		r.t.Logf("resolving link: link=%q jar_c=%#v stack_c=%d stack_h=%d",
			link, len(cookieJar.Cookies(u)), callStack.Count(link), callStack.Height())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return ""
	}

	for k, v := range generateUserAgentHeaders() {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing response body (leaked fd)")
		}
	}()

	if resp.StatusCode > 299 && resp.StatusCode < 400 {
		// We got a redirect
		tu, err := url.Parse(resp.Header.Get("location"))
		if err != nil {
			return ""
		}
		target := r.resolveReference(u, tu)
		callStack.Visit(link)
		return r.resolveFinal(target, cookieJar, callStack)
	}

	// We got a response, it's no redirect, lets check for in-document stuff
	docBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	if metaRedir, err := resolveMetaRedirect(docBody); err == nil {
		// Meta-Redirect found
		tu, err := url.Parse(metaRedir)
		if err != nil {
			return ""
		}
		target := r.resolveReference(u, tu)
		callStack.Visit(link)
		return r.resolveFinal(target, cookieJar, callStack)
	}

	if resp.Header.Get("Set-Cookie") != "" {
		// A new cookie was set, lets refresh the page once to see if stuff
		// changes with that new cookie
		callStack.Visit(link)
		return r.resolveFinal(u.String(), cookieJar, callStack)
	}

	// We had no in-document redirects: we count this as a success
	return u.String()
}

func (resolver) resolveReference(origin *url.URL, loc *url.URL) string {
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

func (r resolver) runResolver() {
	for qe := range r.resolverC {
		if link := r.resolveFinal(qe.Link, r.getJar(), &stack{}); link != "" {
			qe.Callback(link)
		}
		qe.WaitGroup.Done()
	}
}
