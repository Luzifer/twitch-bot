package linkcheck

import (
	"regexp"
	"strings"
	"sync"

	"github.com/Luzifer/go_helpers/v2/str"
)

type (
	// Checker contains logic to detect and resolve links in a message
	Checker struct {
		res *resolver
	}
)

// New creates a new Checker instance with default settings
func New(opts ...func(*Checker)) *Checker {
	c := &Checker{
		res: defaultResolver,
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

func withResolver(r *resolver) func(*Checker) {
	return func(c *Checker) { c.res = r }
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
	wg := new(sync.WaitGroup)

	for ptJoin := 2; ptJoin < len(parts); ptJoin++ {
		for i := 0; i <= len(parts)-ptJoin; i++ {
			c.res.Resolve(resolverQueueEntry{
				Link:      strings.Join(parts[i:i+ptJoin], connector),
				Callback:  func(link string) { links = str.AppendIfMissing(links, link) },
				WaitGroup: wg,
			})
		}
	}

	wg.Wait()

	return links
}

func (c Checker) scanPlainNoObfuscate(message string) (links []string) {
	var (
		parts = regexp.MustCompile(`\s+`).Split(message, -1)
		wg    = new(sync.WaitGroup)
	)

	for _, part := range parts {
		c.res.Resolve(resolverQueueEntry{
			Link:      part,
			Callback:  func(link string) { links = str.AppendIfMissing(links, link) },
			WaitGroup: wg,
		})
	}

	wg.Wait()

	return links
}
