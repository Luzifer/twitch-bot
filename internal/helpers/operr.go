package helpers

import (
	"errors"
	"net"
	"regexp"
)

// CleanOpError checks whether a *net.OpError is included in the error
// and if so removes the included address information. This can happen
// in two ways: If the passed error is indeed an OpError the address
// info is just patched out. If the OpError is buried deeper inside
// the wrapped error stack, a new error with patched message is created
// sacrificing the wrapping and possible included stacktrace.
//
// As of the loss of information this is only intended to clean up
// logging and not be used in error returns.
func CleanOpError(err error) error {
	if opE, ok := err.(*net.OpError); ok {
		// Error in the outmost position is an OpError, lets just patch it
		opE.Source = nil
		opE.Addr = nil
		return opE
	}

	var opE *net.OpError
	if !errors.As(err, &opE) {
		// There is no OpError somewhere inside, keep the error as is
		return err
	}

	// Patch out IP information and create an new error with its message
	return errors.New(regexp.MustCompile(` (?:(?:[0-9]+\.){3}[0-9]+:[0-9]+(?:->)?)+`).
		ReplaceAllString(err.Error(), ""))
}
