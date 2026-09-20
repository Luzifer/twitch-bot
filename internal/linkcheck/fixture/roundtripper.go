// Package fixture contains replicated real-world-behavior of
// websites to test link detection against
package fixture

import (
	"fmt"
	"net/http"
	"regexp"
)

type (
	fixtureGeneratorFunc func(*http.Request) (*http.Response, error)

	fixtureMatch struct {
		Match     *regexp.Regexp
		Generator fixtureGeneratorFunc
	}

	roundTripper struct {
		fixtures []fixtureMatch
	}
)

var fixtures = []fixtureMatch{
	{
		Match:     regexp.MustCompile(`(?i)^https?://archive\.org(?:/|$)`),
		Generator: simpleFixtureResolver(fixture{StatusCode: http.StatusOK}),
	},
	{
		Match:     regexp.MustCompile(`^https://bit\.ly/3KHpJuy$`),
		Generator: headerRedirect(http.StatusMovedPermanently, "https://instagram.com/instagram"),
	},
	{
		Match:     regexp.MustCompile(`^https://bit\.ly/438obkJ$`),
		Generator: headerRedirect(http.StatusMovedPermanently, "https://example.com/"),
	},
	{
		// Claimed by some domain reseller with their generic landingpage
		Match:     regexp.MustCompile(`^https?://btw\.es(?:/|$)`),
		Generator: headerRedirect(http.StatusMovedPermanently, "https://trusted.domainseller.site/btw.es"),
	},
	{
		Match:     regexp.MustCompile(`^https://(?:clips|www)\.twitch\.tv/`),
		Generator: simpleFixtureResolver(fixture{StatusCode: http.StatusOK}),
	},
	{
		Match:     regexp.MustCompile(`^https://cookieguard\.invalid/$`),
		Generator: cookieGuardRedirect("kruemelmonster", "likes-cookies", "https://example.com/"),
	},
	{
		// We use that as a default example which always resolves HTTP 200
		Match:     regexp.MustCompile(`^https?://example\.com(?:/|$)`),
		Generator: simpleFixtureResolver(fixture{StatusCode: http.StatusOK}),
	},
	{
		Match:     regexp.MustCompile(`^http://http521\.invalid/$`),
		Generator: simpleFixtureResolver(fixture{StatusCode: 521}),
	},
	{
		Match:     regexp.MustCompile(`^https://instagram\.com/instagram$`),
		Generator: headerRedirect(http.StatusMovedPermanently, "https://www.instagram.com/instagram"),
	},
	{
		Match:     regexp.MustCompile(`^https://knut\.in/youtube$`),
		Generator: headerRedirect(http.StatusFound, "https://www.youtube.com/channel/UCjsRmaAQ0IHR2CNEBqfNOSQ"),
	},
	{
		Match:     regexp.MustCompile(`^http://metaredirect\.invalid/$`),
		Generator: metaRedirect("https://example.com/"),
	},
	{
		Match:     regexp.MustCompile(`^https://trusted\.domainseller\.site/`),
		Generator: simpleFixtureResolver(fixture{StatusCode: http.StatusOK}),
	},
	{
		Match:     regexp.MustCompile(`^https://twitch\.tv/twitch$`),
		Generator: headerRedirect(http.StatusMovedPermanently, "https://www.twitch.tv/twitch"),
	},
	{
		Match:     regexp.MustCompile(`^https://vk\.cc/hfOwt$`),
		Generator: headerRedirect(http.StatusFound, "https://vkontakte.ru/away.php?cc_key=hfOwt&to=http%3A%2F%2Fyandex.ru%2F"),
	},
	{
		Match:     regexp.MustCompile(`^https://www\.instagram\.com/instagram$`),
		Generator: simpleFixtureResolver(fixture{StatusCode: http.StatusOK}),
	},
	{
		Match:     regexp.MustCompile(`^https://www\.youtube\.com/channel/UCjsRmaAQ0IHR2CNEBqfNOSQ$`),
		Generator: headerRedirect(http.StatusFound, "https://consent.youtube.com/m?continue=https%3A%2F%2Fwww.youtube.com%2Fchannel%2FUCjsRmaAQ0IHR2CNEBqfNOSQ%3Fcbrd%3D1&gl=DE&m=0&pc=yt&cm=2&hl=de&src=1"),
	},
	{
		Match:     regexp.MustCompile(`^https?://yandex\.ru/$`),
		Generator: simpleFixtureResolver(fixture{StatusCode: http.StatusOK}),
	},
	{
		// Last match: Nothing handled the request and it is a HTTP request
		// so we rewrite it to HTTPS and try again.
		Match:     regexp.MustCompile(`^http://`),
		Generator: redirectHTTPS,
	},
}

// New returns a pre-populated set of test-fixtures to test against
func New() http.RoundTripper {
	return roundTripper{
		fixtures: fixtures,
	}
}

func (rt roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	for _, fm := range rt.fixtures {
		if !fm.Match.MatchString(r.URL.String()) || fm.Generator == nil {
			continue
		}

		return fm.Generator(r)
	}

	return nil, fmt.Errorf("no generator found for %q", r.URL.String())
}
