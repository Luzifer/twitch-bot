package linkcheck

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"

	"golang.org/x/net/html"
)

var (
	errNoMetaRedir   = fmt.Errorf("no meta-redir found")
	metaRedirContent = regexp.MustCompile(`^[0-9]+;\s*url=(.*)$`)
)

//nolint:gocognit // Makes no sense to split
func resolveMetaRedirect(body []byte) (redir string, err error) {
	tok := html.NewTokenizer(bytes.NewReader(body))

tokenLoop:
	for {
		token := tok.Next()
		switch token {
		case html.ErrorToken:
			if errors.Is(tok.Err(), io.EOF) {
				break tokenLoop
			}
			return "", fmt.Errorf("scanning tokens: %w", tok.Err())

		case html.StartTagToken:
			t := tok.Token()
			if t.Data == "meta" {
				var (
					content    string
					isRedirect bool
				)

				for _, attr := range t.Attr {
					isRedirect = isRedirect || attr.Key == "http-equiv" && attr.Val == "refresh"

					if attr.Key == "content" {
						content = attr.Val
					}
				}

				if !isRedirect {
					continue tokenLoop
				}

				// It is a redirect, get the content and parse it
				if matches := metaRedirContent.FindStringSubmatch(content); len(matches) > 1 {
					redir = matches[1]
				}
			}
		}
	}

	if redir == "" {
		// We did not find anything
		return "", errNoMetaRedir
	}

	return redir, nil
}
