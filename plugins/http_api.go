package plugins

import "net/http"

type (
	// HTTPRouteParamDocumentation documents parameters expected by a
	// HTTP route and to be documented in the API documentation
	HTTPRouteParamDocumentation struct {
		Description string
		Name        string
		Required    bool
		Type        string
	}

	// HTTPRouteRegistrationArgs defines the HTTP route to be added in
	// using the HTTPRouteRegistrationFunc
	HTTPRouteRegistrationArgs struct {
		Accept              []string
		Description         string
		HandlerFunc         http.HandlerFunc
		IsPrefix            bool
		Method              string
		Module              string
		Name                string
		Path                string
		QueryParams         []HTTPRouteParamDocumentation
		RequiresEditorsAuth bool
		RequiresWriteAuth   bool
		ResponseType        HTTPRouteResponseType
		RouteParams         []HTTPRouteParamDocumentation
		SkipDocumentation   bool
	}

	// HTTPRouteResponseType pre-defines response types known to the API
	// documentation
	HTTPRouteResponseType uint64

	// HTTPRouteRegistrationFunc is passed from the bot to the
	// plugins RegisterFunc to register a new route in the API router
	HTTPRouteRegistrationFunc func(HTTPRouteRegistrationArgs) error
)

// Enum of known HTTPRouteResponseType
const (
	HTTPRouteResponseTypeNo200 HTTPRouteResponseType = iota
	HTTPRouteResponseTypeTextPlain
	HTTPRouteResponseTypeJSON
	HTTPRouteResponseTypeMultiple
)
