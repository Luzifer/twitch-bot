package helpers

import (
	"errors"
	"net"
	"regexp"
)

var networkArrowErrorPart = regexp.MustCompile(` (?:(?:[0-9]+\.){3}[0-9]+:[0-9]+(?:->)?)+`)

// CleanNetworkAddressFromError checks whether an IP:Port->IP:port
// information is contained in the error. This is checked by explicitly
// sanitizing *net.OpError instances or by returning a sanitized error
// string without the stack previously present.
//
// As of the loss of information this is only intended to clean up
// logging and not be used in error returns.
func CleanNetworkAddressFromError(err error) error {
	if opE, ok := err.(*net.OpError); ok {
		// Error in the outmost position is an OpError, lets just patch it
		opE.Source = nil
		opE.Addr = nil
		return opE
	}

	if networkArrowErrorPart.FindStringIndex(err.Error()) == nil {
		// There is no network address somewhere inside, keep the error as is
		return err
	}

	// Patch out IP information and create an new error with its message
	return errors.New(networkArrowErrorPart.ReplaceAllString(err.Error(), ""))
}
