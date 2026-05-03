package raffle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const moduleName = "raffle"

var apiRoutes = []plugins.HTTPRouteRegistrationArgs{
	{
		Description: "Lists all raffles known to the bot",
		HandlerFunc: handleWrap(func(_ http.ResponseWriter, _ *http.Request, _ map[string]uint64) (any, error) {
			ras, err := dbc.List()
			if err != nil {
				return ras, fmt.Errorf("fetching raffles from database: %w", err)
			}
			return ras, nil
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
		HandlerFunc: handleWrap(func(_ http.ResponseWriter, r *http.Request, _ map[string]uint64) (any, error) {
			var ra raffle
			if err := json.NewDecoder(r.Body).Decode(&ra); err != nil {
				return nil, fmt.Errorf("parsing raffle from body: %w", err)
			}

			if err := dbc.Create(ra); err != nil {
				return nil, fmt.Errorf("creating raffle: %w", err)
			}

			return nil, nil //nolint:nilnil // This is no error but the fine state, we're just matching the interface
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
		HandlerFunc: handleWrap(func(_ http.ResponseWriter, _ *http.Request, ids map[string]uint64) (any, error) {
			if err := dbc.Delete(ids["id"]); err != nil {
				return nil, fmt.Errorf("fetching raffle from database: %w", err)
			}

			return nil, nil //nolint:nilnil // This is no error but the fine state, we're just matching the interface
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
		HandlerFunc: handleWrap(func(_ http.ResponseWriter, _ *http.Request, ids map[string]uint64) (any, error) {
			ra, err := dbc.Get(ids["id"])
			if err != nil {
				return ra, fmt.Errorf("fetching raffle from database: %w", err)
			}
			return ra, nil
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
		HandlerFunc: handleWrap(func(_ http.ResponseWriter, r *http.Request, ids map[string]uint64) (any, error) {
			var ra raffle
			if err := json.NewDecoder(r.Body).Decode(&ra); err != nil {
				return nil, fmt.Errorf("parsing raffle from body: %w", err)
			}

			if ra.ID != ids["id"] {
				return nil, fmt.Errorf("raffle ID does not match")
			}

			if err := dbc.Update(ra); err != nil {
				return nil, fmt.Errorf("updating raffle: %w", err)
			}

			return nil, nil //nolint:nilnil // This is no error but the fine state, we're just matching the interface
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
		Description: "Resets the raffle (remove entries, reset status & start/close time) given by its ID",
		HandlerFunc: handleWrap(func(_ http.ResponseWriter, _ *http.Request, ids map[string]uint64) (any, error) {
			if err := dbc.Reset(ids["id"]); err != nil {
				return nil, fmt.Errorf("resetting raffle: %w", err)
			}

			return nil, nil //nolint:nilnil // This is no error but the fine state, we're just matching the interface
		}, []string{"id"}),
		Method:            http.MethodPut,
		Module:            moduleName,
		Name:              "Reset Raffle",
		Path:              "/{id}/reset",
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeNo200,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "ID of the raffle to reset",
				Name:        "id",
			},
		},
	},

	{
		Description: "Duplicates the raffle given by its ID",
		HandlerFunc: handleWrap(func(_ http.ResponseWriter, _ *http.Request, ids map[string]uint64) (any, error) {
			if err := dbc.Clone(ids["id"]); err != nil {
				return nil, fmt.Errorf("cloning raffle: %w", err)
			}

			return nil, nil //nolint:nilnil // This is no error but the fine state, we're just matching the interface
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
		HandlerFunc: handleWrap(func(_ http.ResponseWriter, _ *http.Request, ids map[string]uint64) (any, error) {
			if err := dbc.Close(ids["id"]); err != nil {
				return nil, fmt.Errorf("closing raffle: %w", err)
			}

			return nil, nil //nolint:nilnil // This is no error but the fine state, we're just matching the interface
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
		HandlerFunc: handleWrap(func(_ http.ResponseWriter, _ *http.Request, ids map[string]uint64) (any, error) {
			if err := dbc.PickWinner(ids["id"]); err != nil {
				return nil, fmt.Errorf("picking winner: %w", err)
			}

			return nil, nil //nolint:nilnil // This is no error but the fine state, we're just matching the interface
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
		HandlerFunc: handleWrap(func(_ http.ResponseWriter, r *http.Request, ids map[string]uint64) (any, error) {
			dur, err := strconv.ParseInt(r.URL.Query().Get("duration"), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parsing duration: %w", err)
			}

			if err := dbc.Reopen(ids["id"], time.Duration(dur)*time.Second); err != nil {
				return nil, fmt.Errorf("reopening raffle: %w", err)
			}

			return nil, nil //nolint:nilnil // This is no error but the fine state, we're just matching the interface
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
		HandlerFunc: handleWrap(func(_ http.ResponseWriter, _ *http.Request, ids map[string]uint64) (any, error) {
			if err := dbc.Start(ids["id"]); err != nil {
				return nil, fmt.Errorf("starting raffle: %w", err)
			}

			return nil, nil //nolint:nilnil // This is no error but the fine state, we're just matching the interface
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
		HandlerFunc: handleWrap(func(_ http.ResponseWriter, _ *http.Request, ids map[string]uint64) (any, error) {
			if err := dbc.RedrawWinner(ids["id"], ids["winner"]); err != nil {
				return nil, fmt.Errorf("re-picking winner: %w", err)
			}

			return nil, nil //nolint:nilnil // This is no error but the fine state, we're just matching the interface
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
			ids     = make(map[string]uint64)
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

		w.Header().Set("Content-Type", "application/json")
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
			return fmt.Errorf("registering route %d: %w", i, err)
		}
	}

	return nil
}
