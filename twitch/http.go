package twitch

import (
	"fmt"
)

type httpError struct {
	body []byte
	code int
	err  error
}

var errAnyHTTPError = newHTTPError(0, nil, nil)

func newHTTPError(status int, body []byte, wrappedErr error) httpError {
	return httpError{
		body: body,
		code: status,
		err:  wrappedErr,
	}
}

func (h httpError) Error() string {
	selfE := fmt.Sprintf("unexpected status %d", h.code)
	if h.body != nil {
		selfE = fmt.Sprintf("%s (%s)", selfE, h.body)
	}

	if h.err == nil {
		return selfE
	}

	return fmt.Sprintf("%s: %s", selfE, h.err)
}

func (h httpError) Is(target error) bool {
	ht, ok := target.(httpError)
	if !ok {
		return false
	}

	return ht.code == 0 || ht.code == h.code
}

func (h httpError) Unwrap() error {
	return h.err
}
