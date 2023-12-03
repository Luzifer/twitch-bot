package raffle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const moduleName = "raffle"

var apiRoutes = []plugins.HTTPRouteRegistrationArgs{
	{
		Description: "Lists all raffles known to the bot",
		HandlerFunc: handleWrap(func(w http.ResponseWriter, r *http.Request, _ map[string]uint64) (any, error) {
			ras, err := dbc.List()
			return ras, errors.Wrap(err, "fetching raffles from database")
		}, nil),
		Method:            http.MethodGet,
		Module:            moduleName,
		Name:              "List Raffles",
		Path:              "/",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeJSON,
	},

	{
		Description: "Creates a new raffle based on the data in the body",
		HandlerFunc: handleWrap(func(w http.ResponseWriter, r *http.Request, _ map[string]uint64) (any, error) {
			var ra raffle
			if err := json.NewDecoder(r.Body).Decode(&ra); err != nil {
				return nil, errors.Wrap(err, "parsing raffle from body")
			}

			return nil, errors.Wrap(dbc.Create(ra), "creating raffle")
		}, nil),
		Method:            http.MethodPost,
		Module:            moduleName,
		Name:              "Create Raffle",
		Path:              "/",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeNo200,
	},

	{
		Description: "Deletes raffle by given ID including all entries",
		HandlerFunc: handleWrap(func(w http.ResponseWriter, r *http.Request, ids map[string]uint64) (any, error) {
			return nil, errors.Wrap(dbc.Delete(ids["id"]), "fetching raffle from database")
		}, []string{"id"}),
		Method:            http.MethodDelete,
		Module:            moduleName,
		Name:              "Delete Raffle",
		Path:              "/{id}",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeNo200,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "ID of the raffle to fetch",
				Name:        "id",
			},
		},
	},

	{
		Description: "Gets raffle by given ID including all entries",
		HandlerFunc: handleWrap(func(w http.ResponseWriter, r *http.Request, ids map[string]uint64) (any, error) {
			ra, err := dbc.Get(ids["id"])
			return ra, errors.Wrap(err, "fetching raffle from database")
		}, []string{"id"}),
		Method:            http.MethodGet,
		Module:            moduleName,
		Name:              "Get Raffle",
		Path:              "/{id}",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeJSON,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "ID of the raffle to fetch",
				Name:        "id",
			},
		},
	},

	{
		Description: "Updates the given raffle (needs to include the whole object, not just changed fields)",
		HandlerFunc: handleWrap(func(w http.ResponseWriter, r *http.Request, ids map[string]uint64) (any, error) {
			var ra raffle
			if err := json.NewDecoder(r.Body).Decode(&ra); err != nil {
				return nil, errors.Wrap(err, "parsing raffle from body")
			}

			if ra.ID != ids["id"] {
				return nil, errors.New("raffle ID does not match")
			}

			return nil, errors.Wrap(dbc.Update(ra), "updating raffle")
		}, []string{"id"}),
		Method:            http.MethodPut,
		Module:            moduleName,
		Name:              "Update Raffle",
		Path:              "/{id}",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeNo200,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "ID of the raffle to update",
				Name:        "id",
			},
		},
	},

	{
		Description: "Duplicates the raffle given by its ID",
		HandlerFunc: handleWrap(func(w http.ResponseWriter, r *http.Request, ids map[string]uint64) (any, error) {
			return nil, errors.Wrap(dbc.Clone(ids["id"]), "cloning raffle")
		}, []string{"id"}),
		Method:            http.MethodPut,
		Module:            moduleName,
		Name:              "Clone Raffle",
		Path:              "/{id}/clone",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeNo200,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "ID of the raffle to clone",
				Name:        "id",
			},
		},
	},

	{
		Description: "Closes the raffle given by its ID",
		HandlerFunc: handleWrap(func(w http.ResponseWriter, r *http.Request, ids map[string]uint64) (any, error) {
			return nil, errors.Wrap(dbc.Close(ids["id"]), "closing raffle")
		}, []string{"id"}),
		Method:            http.MethodPut,
		Module:            moduleName,
		Name:              "Close Raffle",
		Path:              "/{id}/close",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeNo200,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "ID of the raffle to close",
				Name:        "id",
			},
		},
	},

	{
		Description: "Picks a winner for the given raffle (this does NOT close the raffle, use only on closed raffle!)",
		HandlerFunc: handleWrap(func(w http.ResponseWriter, r *http.Request, ids map[string]uint64) (any, error) {
			return nil, errors.Wrap(dbc.PickWinner(ids["id"]), "picking winner")
		}, []string{"id"}),
		Method:            http.MethodPut,
		Module:            moduleName,
		Name:              "Pick Raffle Winner",
		Path:              "/{id}/pick",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeNo200,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "ID of the raffle to pick a winner for",
				Name:        "id",
			},
		},
	},

	{
		Description: "Re-opens a raffle for additional entries, only Status and CloseAt are modified",
		HandlerFunc: handleWrap(func(w http.ResponseWriter, r *http.Request, ids map[string]uint64) (any, error) {
			dur, err := strconv.ParseInt(r.URL.Query().Get("duration"), 10, 64)
			if err != nil {
				return nil, errors.Wrap(err, "parsing duration")
			}

			return nil, errors.Wrap(dbc.Reopen(ids["id"], time.Duration(dur)*time.Second), "reopening raffle")
		}, []string{"id"}),
		Method: http.MethodPut,
		Module: moduleName,
		Name:   "Reopen Raffle",
		Path:   "/{id}/reopen",
		QueryParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Number of seconds to leave the raffle open",
				Name:        "duration",
				Required:    true,
				Type:        "int",
			},
		},
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeNo200,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "ID of the raffle to pick a winner for",
				Name:        "id",
			},
		},
	},

	{
		Description: "Starts a raffle making it available for entries",
		HandlerFunc: handleWrap(func(w http.ResponseWriter, r *http.Request, ids map[string]uint64) (any, error) {
			return nil, errors.Wrap(dbc.Start(ids["id"]), "starting raffle")
		}, []string{"id"}),
		Method:            http.MethodPut,
		Module:            moduleName,
		Name:              "Start Raffle",
		Path:              "/{id}/start",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeNo200,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "ID of the raffle to start",
				Name:        "id",
			},
		},
	},

	{
		Description: "Dismisses a previously picked winner and picks a new one",
		HandlerFunc: handleWrap(func(w http.ResponseWriter, r *http.Request, ids map[string]uint64) (any, error) {
			return nil, errors.Wrap(dbc.RedrawWinner(ids["id"], ids["winner"]), "re-picking winner")
		}, []string{"id", "winner"}),
		Method:            http.MethodPut,
		Module:            moduleName,
		Name:              "Re-Pick Raffle Winner",
		Path:              "/{id}/repick/{winner}",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeNo200,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "ID of the raffle to re-pick the winner for",
				Name:        "id",
			},
			{
				Description: "ID of the winner to replace",
				Name:        "winner",
			},
		},
	},
}

func handleWrap(f func(http.ResponseWriter, *http.Request, map[string]uint64) (any, error), parseIDs []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ids     = map[string]uint64{}
			reqUUID = uuid.Must(uuid.NewV4()).String()
			logger  = logrus.WithFields(logrus.Fields{
				"path": r.URL.Path,
				"req":  reqUUID,
			})
		)

		for _, k := range parseIDs {
			id, err := strconv.ParseUint(mux.Vars(r)[k], 10, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid ID field %q", k), http.StatusBadRequest)
				return
			}

			ids[k] = id
		}

		resp, err := f(w, r, ids)
		if err != nil {
			logger.WithError(err).Error("handling request")
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		if resp == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "applicatioh/json")
		if err = json.NewEncoder(w).Encode(resp); err != nil {
			logger.WithError(err).Error("serializing response")
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}
	}
}

func registerAPI(args plugins.RegistrationArguments) (err error) {
	for i, r := range apiRoutes {
		if err = args.RegisterAPIRoute(r); err != nil {
			return errors.Wrapf(err, "registering route %d", i)
		}
	}

	return nil
}
