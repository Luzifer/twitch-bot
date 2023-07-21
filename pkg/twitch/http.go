package twitch

import (
	"fmt"
)

// HTTPError represents an HTTP error containing the response body (or
// the wrapped error occurred while readiny the body) and the status
// code returned by the server
type HTTPError struct {
	Body []byte
	Code int
	Err  error
}

// ErrAnyHTTPError can be used in errors.Is() to match an HTTPError
// with any status code
var ErrAnyHTTPError = newHTTPError(0, nil, nil)

func newHTTPError(status int, body []byte, wrappedErr error) HTTPError {
	return HTTPError{
		Body: body,
		Code: status,
		Err:  wrappedErr,
	}
}

// Error implements the error interface and returns a formatted version
// of the error including the body, might therefore leak confidential
// information when included in the response body
func (h HTTPError) Error() string {
	selfE := fmt.Sprintf("unexpected status %d", h.Code)
	if h.Body != nil {
		selfE = fmt.Sprintf("%s (%s)", selfE, h.Body)
	}

	if h.Err == nil {
		return selfE
	}

	return fmt.Sprintf("%s: %s", selfE, h.Err)
}

// Is checks whether the given error is an HTTPError and the status
// code matches the given error
func (h HTTPError) Is(target error) bool {
	ht, ok := target.(HTTPError)
	if !ok {
		return false
	}

	return ht.Code == 0 || ht.Code == h.Code
}

// Unwrap returns the wrapped error occurred when reading the body
func (h HTTPError) Unwrap() error {
	return h.Err
}
