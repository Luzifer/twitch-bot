package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/wzshiming/openapi/spec"
)

var (
	swaggerDoc = spec.OpenAPI{
		OpenAPI: "3.0.3", // This generator uses v3 of OpenAPI standard
		Info: &spec.Info{
			Title:   "Twitch-Bot public API",
			Version: "v1",
		},
		Servers: []*spec.Server{
			{URL: "./", Description: "Current bot instance"},
		},
		Paths: make(spec.Paths),
		Components: &spec.Components{
			Responses: map[string]*spec.Response{
				"genericErrorResponse": spec.TextPlainResponse(nil).WithDescription("An error occurred: See error message"),
				"inputErrorResponse":   spec.TextPlainResponse(nil).WithDescription("Data sent to API is invalid: See error message"),
				"notFoundResponse":     spec.TextPlainResponse(nil).WithDescription("Document was not found or insufficient permissions"),
			},
			SecuritySchemes: map[string]*spec.SecurityScheme{
				"authenticated": spec.APIKeyAuth("Authorization", spec.InHeader),
			},
		},
	}

	//go:embed swagger.html
	swaggerHTML []byte
)

func handleSwaggerHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	if _, err := io.Copy(w, bytes.NewReader(swaggerHTML)); err != nil {
		http.Error(w, errors.Wrap(err, "writing frontend").Error(), http.StatusInternalServerError)
	}
}

func handleSwaggerRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(swaggerDoc); err != nil {
		http.Error(w, errors.Wrap(err, "rendering documentation").Error(), http.StatusInternalServerError)
	}
}

//nolint:gocyclo // Makes no sense to split just to spare a little complexity
func registerSwaggerRoute(route plugins.HTTPRouteRegistrationArgs) error {
	fullPath := strings.Join([]string{
		"",
		route.Module,
		strings.TrimLeft(route.Path, "/"),
	}, "/")

	pi, ok := swaggerDoc.Paths[fullPath]
	if !ok {
		pi = &spec.PathItem{}

		for _, param := range route.RouteParams {
			pi.Parameters = append(
				pi.Parameters,
				spec.PathParam(param.Name, spec.StringProperty()).WithDescription(param.Description),
			)
		}

		swaggerDoc.Paths[fullPath] = pi
	}

	op := &spec.Operation{
		Summary:     route.Name,
		Description: route.Description,
		Tags:        []string{route.Module},
		Responses: map[string]*spec.Response{
			"204": spec.TextPlainResponse(nil).WithDescription("Successful execution without response object"),
			"404": spec.RefResponse("notFoundResponse"),
			"500": spec.RefResponse("genericErrorResponse"),
		},
	}

	if route.RequiresEditorsAuth {
		op.Security = []map[string]spec.SecurityRequirement{
			{"authenticated": {}},
		}
	}

	switch route.ResponseType {
	case plugins.HTTPRouteResponseTypeJSON:
		op.Responses["200"] = spec.JSONResponse(nil).WithDescription("Successful execution with JSON object response")

	case plugins.HTTPRouteResponseTypeMultiple:
		op.Responses["200"] = (&spec.Response{}).WithDescription("Successful execution with variable response based on Accept header")
		for _, a := range route.Accept {
			op.Responses["200"].AddContent(a, &spec.MediaType{Schema: nil})
		}

	case plugins.HTTPRouteResponseTypeNo200:
		// We don't add a 200 then

	case plugins.HTTPRouteResponseTypeTextPlain:
		op.Responses["200"] = spec.TextPlainResponse(nil).WithDescription("Successful execution with plain text response")
	}

	for _, param := range route.QueryParams {
		var ps *spec.Schema

		switch param.Type {
		case "bool", "boolean":
			ps = spec.BooleanProperty()

		case "int", "int64":
			ps = spec.Int64Property()

		case "string":
			ps = spec.StringProperty()

		default:
			log.WithFields(log.Fields{"module": route.Module, "type": param.Type}).Warn("Module registered unhandled query-param type")
			ps = spec.StringProperty()
		}

		specParam := spec.QueryParam(param.Name, ps).
			WithDescription(param.Description)

		specParam.Required = param.Required

		op.Parameters = append(
			op.Parameters,
			specParam,
		)
	}

	switch route.Method {
	case http.MethodDelete:
		pi.Delete = op
	case http.MethodGet:
		pi.Get = op
	case http.MethodPatch:
		op.Responses["400"] = spec.RefResponse("inputErrorResponse")
		pi.Patch = op
	case http.MethodPost:
		op.Responses["400"] = spec.RefResponse("inputErrorResponse")
		pi.Post = op
	case http.MethodPut:
		op.Responses["400"] = spec.RefResponse("inputErrorResponse")
		pi.Put = op
	default:
		return errors.Errorf("assignment for %q is not implemented", route.Method)
	}

	return nil
}
