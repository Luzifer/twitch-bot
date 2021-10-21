package quotedb

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func registerAPI(register plugins.HTTPRouteRegistrationFunc) {
	register(plugins.HTTPRouteRegistrationArgs{
		Description:  "Add quotes for the given {channel}",
		HandlerFunc:  handleAddQuotes,
		Method:       http.MethodPost,
		Module:       "quotedb",
		Name:         "Add Quotes",
		Path:         "/{channel}",
		ResponseType: plugins.HTTPRouteResponseTypeTextPlain,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Channel to delete the quote in",
				Name:        "channel",
			},
		},
	})

	register(plugins.HTTPRouteRegistrationArgs{
		Description:  "Deletes quote with given {idx} from {channel}",
		HandlerFunc:  handleDeleteQuote,
		Method:       http.MethodDelete,
		Module:       "quotedb",
		Name:         "Delete Quote",
		Path:         "/{channel}/{idx:[0-9]+}",
		ResponseType: plugins.HTTPRouteResponseTypeTextPlain,
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
	})

	register(plugins.HTTPRouteRegistrationArgs{
		Description:  "Lists all quotes for the given {channel}",
		HandlerFunc:  handleListQuotes,
		Method:       http.MethodGet,
		Module:       "quotedb",
		Name:         "List Quotes",
		Path:         "/{channel}",
		ResponseType: plugins.HTTPRouteResponseTypeJSON,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Channel to delete the quote in",
				Name:        "channel",
			},
		},
	})

	register(plugins.HTTPRouteRegistrationArgs{
		Description:  "Set quotes for the given {channel} (will overwrite ALL quotes!)",
		HandlerFunc:  handleReplaceQuotes,
		Method:       http.MethodPut,
		Module:       "quotedb",
		Name:         "Set Quotes",
		Path:         "/{channel}",
		ResponseType: plugins.HTTPRouteResponseTypeTextPlain,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Channel to delete the quote in",
				Name:        "channel",
			},
		},
	})

	register(plugins.HTTPRouteRegistrationArgs{
		Description:  "Updates quote with given {idx} from {channel}",
		HandlerFunc:  handleDeleteQuote,
		Method:       http.MethodPut,
		Module:       "quotedb",
		Name:         "Update Quote",
		Path:         "/{channel}/{idx:[0-9]+}",
		ResponseType: plugins.HTTPRouteResponseTypeTextPlain,
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
	})
}

func handleAddQuotes(w http.ResponseWriter, r *http.Request) {
	channel := "#" + strings.TrimLeft(mux.Vars(r)["channel"], "#")

	var quotes []string
	if err := json.NewDecoder(r.Body).Decode(&quotes); err != nil {
		http.Error(w, errors.Wrap(err, "parsing input").Error(), http.StatusBadRequest)
		return
	}

	for _, q := range quotes {
		storedObject.AddQuote(channel, q)
	}

	if err := store.SetModuleStore(moduleUUID, storedObject); err != nil {
		http.Error(w, errors.Wrap(err, "storing quote database").Error(), http.StatusInternalServerError)
		return
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

	storedObject.DelQuote(channel, idx)

	if err := store.SetModuleStore(moduleUUID, storedObject); err != nil {
		http.Error(w, errors.Wrap(err, "storing quote database").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleListQuotes(w http.ResponseWriter, r *http.Request) {
	channel := "#" + strings.TrimLeft(mux.Vars(r)["channel"], "#")

	quotes := storedObject.GetChannelQuotes(channel)

	if err := json.NewEncoder(w).Encode(quotes); err != nil {
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

	storedObject.SetQuotes(channel, quotes)

	if err := store.SetModuleStore(moduleUUID, storedObject); err != nil {
		http.Error(w, errors.Wrap(err, "storing quote database").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

	storedObject.UpdateQuote(channel, idx, quotes[0])

	if err := store.SetModuleStore(moduleUUID, storedObject); err != nil {
		http.Error(w, errors.Wrap(err, "storing quote database").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
