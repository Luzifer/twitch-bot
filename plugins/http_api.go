package plugins

import "net/http"

type (
	HTTPRouteParamDocumentation struct {
		Description string
		Name        string
		Required    bool
		Type        string
	}

	HTTPRouteRegistrationArgs struct {
		Description       string
		HandlerFunc       http.HandlerFunc
		IsPrefix          bool
		Method            string
		Module            string
		Name              string
		Path              string
		QueryParams       []HTTPRouteParamDocumentation
		ResponseType      HTTPRouteResponseType
		RouteParams       []HTTPRouteParamDocumentation
		SkipDocumentation bool
	}

	HTTPRouteResponseType uint64

	HTTPRouteRegistrationFunc func(HTTPRouteRegistrationArgs) error
)

const (
	HTTPRouteResponseTypeNo200 HTTPRouteResponseType = iota
	HTTPRouteResponseTypeTextPlain
	HTTPRouteResponseTypeJSON
)
