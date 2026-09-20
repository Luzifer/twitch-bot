package fixture

import (
	"io"
	"net/http"
	"strings"
)

type (
	fixture struct {
		Body       string
		Header     http.Header
		StatusCode int
	}
)

func simpleFixtureResolver(f fixture) fixtureGeneratorFunc {
	return func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			Body:       io.NopCloser(strings.NewReader(f.Body)),
			Header:     f.Header.Clone(),
			Request:    r,
			StatusCode: f.StatusCode,
		}, nil
	}
}
