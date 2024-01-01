package quotedb

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var (
	//go:embed list.html
	listFrontend []byte
	//go:embed list.js
	listScript []byte
)

//nolint:funlen
func registerAPI(register plugins.HTTPRouteRegistrationFunc) (err error) {
	if err = register(plugins.HTTPRouteRegistrationArgs{
		HandlerFunc:       handleScript,
		Method:            http.MethodGet,
		Module:            "quotedb",
		Path:              "/app.js",
		SkipDocumentation: true,
	}); err != nil {
		return fmt.Errorf("registering API route: %w", err)
	}

	if err = register(plugins.HTTPRouteRegistrationArgs{
		Description:       "Add quotes for the given {channel}",
		HandlerFunc:       handleAddQuotes,
		Method:            http.MethodPost,
		Module:            "quotedb",
		Name:              "Add Quotes",
		Path:              "/{channel}",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeTextPlain,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Channel to delete the quote in",
				Name:        "channel",
			},
		},
	}); err != nil {
		return fmt.Errorf("registering API route: %w", err)
	}

	if err = register(plugins.HTTPRouteRegistrationArgs{
		Description:       "Deletes quote with given {idx} from {channel}",
		HandlerFunc:       handleDeleteQuote,
		Method:            http.MethodDelete,
		Module:            "quotedb",
		Name:              "Delete Quote",
		Path:              "/{channel}/{idx:[0-9]+}",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeTextPlain,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Channel to delete the quote in",
				Name:        "channel",
			},
			{
				Description: "Index of the quote to delete",
				Name:        "idx",
			},
		},
	}); err != nil {
		return fmt.Errorf("registering API route: %w", err)
	}

	if err = register(plugins.HTTPRouteRegistrationArgs{
		Accept:       []string{"application/json", "text/html"},
		Description:  "Lists all quotes for the given {channel}",
		HandlerFunc:  handleListQuotes,
		Method:       http.MethodGet,
		Module:       "quotedb",
		Name:         "List Quotes",
		Path:         "/{channel}",
		ResponseType: plugins.HTTPRouteResponseTypeMultiple,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Channel to delete the quote in",
				Name:        "channel",
			},
		},
	}); err != nil {
		return fmt.Errorf("registering API route: %w", err)
	}

	if err = register(plugins.HTTPRouteRegistrationArgs{
		Description:       "Set quotes for the given {channel} (will overwrite ALL quotes!)",
		HandlerFunc:       handleReplaceQuotes,
		Method:            http.MethodPut,
		Module:            "quotedb",
		Name:              "Set Quotes",
		Path:              "/{channel}",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeTextPlain,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Channel to delete the quote in",
				Name:        "channel",
			},
		},
	}); err != nil {
		return fmt.Errorf("registering API route: %w", err)
	}

	if err = register(plugins.HTTPRouteRegistrationArgs{
		Description:       "Updates quote with given {idx} from {channel}",
		HandlerFunc:       handleUpdateQuote,
		Method:            http.MethodPut,
		Module:            "quotedb",
		Name:              "Update Quote",
		Path:              "/{channel}/{idx:[0-9]+}",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeTextPlain,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Channel to delete the quote in",
				Name:        "channel",
			},
			{
				Description: "Index of the quote to delete",
				Name:        "idx",
			},
		},
	}); err != nil {
		return fmt.Errorf("registering API route: %w", err)
	}

	return nil
}

func handleAddQuotes(w http.ResponseWriter, r *http.Request) {
	channel := "#" + strings.TrimLeft(mux.Vars(r)["channel"], "#")

	var quotes []string
	if err := json.NewDecoder(r.Body).Decode(&quotes); err != nil {
		http.Error(w, errors.Wrap(err, "parsing input").Error(), http.StatusBadRequest)
		return
	}

	for _, q := range quotes {
		if err := addQuote(db, channel, q); err != nil {
			http.Error(w, errors.Wrap(err, "adding quote").Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func handleDeleteQuote(w http.ResponseWriter, r *http.Request) {
	var (
		channel = "#" + strings.TrimLeft(mux.Vars(r)["channel"], "#")
		idxStr  = mux.Vars(r)["idx"]
	)

	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		http.Error(w, errors.Wrap(err, "parsing index").Error(), http.StatusBadRequest)
		return
	}

	if err = delQuote(db, channel, idx); err != nil {
		http.Error(w, errors.Wrap(err, "deleting quote").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleListQuotes(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.Header.Get("Accept"), "text/html") {
		w.Header().Set("Content-Type", "text/html")
		w.Write(listFrontend) //nolint:errcheck,gosec,revive
		return
	}

	channel := "#" + strings.TrimLeft(mux.Vars(r)["channel"], "#")

	quotes, err := getChannelQuotes(db, channel)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting quotes").Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(quotes); err != nil {
		http.Error(w, errors.Wrap(err, "enocding quote list").Error(), http.StatusInternalServerError)
		return
	}
}

func handleReplaceQuotes(w http.ResponseWriter, r *http.Request) {
	channel := "#" + strings.TrimLeft(mux.Vars(r)["channel"], "#")

	var quotes []string
	if err := json.NewDecoder(r.Body).Decode(&quotes); err != nil {
		http.Error(w, errors.Wrap(err, "parsing input").Error(), http.StatusBadRequest)
		return
	}

	if err := setQuotes(db, channel, quotes); err != nil {
		http.Error(w, errors.Wrap(err, "replacing quotes").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleScript(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/javascript")
	w.Write(listScript) //nolint:errcheck,gosec,revive
}

func handleUpdateQuote(w http.ResponseWriter, r *http.Request) {
	var (
		channel = "#" + strings.TrimLeft(mux.Vars(r)["channel"], "#")
		idxStr  = mux.Vars(r)["idx"]
	)

	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		http.Error(w, errors.Wrap(err, "parsing index").Error(), http.StatusBadRequest)
		return
	}

	var quotes []string
	if err := json.NewDecoder(r.Body).Decode(&quotes); err != nil {
		http.Error(w, errors.Wrap(err, "parsing input").Error(), http.StatusBadRequest)
		return
	}

	if len(quotes) != 1 {
		http.Error(w, "input must be list with one quote", http.StatusBadRequest)
		return
	}

	if err = updateQuote(db, channel, idx, quotes[0]); err != nil {
		http.Error(w, errors.Wrap(err, "updating quote").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
