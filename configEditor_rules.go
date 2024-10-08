package main

import (
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid/v3"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

func registerEditorRulesRoutes() {
	for _, rd := range []plugins.HTTPRouteRegistrationArgs{
		{
			Description:         "Returns the current set of configured rules in JSON format",
			HandlerFunc:         configEditorRulesGet,
			Method:              http.MethodGet,
			Module:              moduleConfigEditor,
			Name:                "Get current rules",
			Path:                "/rules",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeJSON,
		},
		{
			Description:         "Adds a new Rule",
			HandlerFunc:         configEditorRulesAdd,
			Method:              http.MethodPost,
			Module:              moduleConfigEditor,
			Name:                "Add Rule",
			Path:                "/rules",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeTextPlain,
		},
		{
			Description:         "Deletes the given Rule",
			HandlerFunc:         configEditorRulesDelete,
			Method:              http.MethodDelete,
			Module:              moduleConfigEditor,
			Name:                "Delete Rule",
			Path:                "/rules/{uuid}",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeTextPlain,
			RouteParams: []plugins.HTTPRouteParamDocumentation{
				{
					Description: "UUID of the rule to delete",
					Name:        "uuid",
					Required:    false,
					Type:        "string",
				},
			},
		},
		{
			Description:         "Updates the given Rule",
			HandlerFunc:         configEditorRulesUpdate,
			Method:              http.MethodPut,
			Module:              moduleConfigEditor,
			Name:                "Update Rule",
			Path:                "/rules/{uuid}",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeTextPlain,
			RouteParams: []plugins.HTTPRouteParamDocumentation{
				{
					Description: "UUID of the rule to update",
					Name:        "uuid",
					Required:    false,
					Type:        "string",
				},
			},
		},
	} {
		if err := registerRoute(rd); err != nil {
			log.WithError(err).Fatal("Unable to register config editor route")
		}
	}
}

func configEditorRulesAdd(w http.ResponseWriter, r *http.Request) {
	user, _, err := getAuthorizationFromRequest(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting authorized user").Error(), http.StatusInternalServerError)
		return
	}

	msg := &plugins.Rule{}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if msg.SubscribeFrom != nil {
		if _, err = msg.UpdateFromSubscription(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		msg.UUID = uuid.Must(uuid.NewV4()).String()
	}

	if err := patchConfig(cfg.Config, user, "", "Add rule", func(c *configFile) error {
		for _, r := range c.Rules {
			if r.UUID == msg.UUID {
				return errors.New("rule already exists (UUID duplicate)")
			}
		}

		c.Rules = append(c.Rules, msg)
		return nil
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func configEditorRulesDelete(w http.ResponseWriter, r *http.Request) {
	user, _, err := getAuthorizationFromRequest(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting authorized user").Error(), http.StatusInternalServerError)
		return
	}

	if err := patchConfig(cfg.Config, user, "", "Delete rule", func(c *configFile) error {
		var (
			id  = mux.Vars(r)["uuid"]
			tmp []*plugins.Rule
		)

		for i := range c.Rules {
			if c.Rules[i].MatcherID() == id {
				continue
			}
			tmp = append(tmp, c.Rules[i])
		}

		c.Rules = tmp

		return nil
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func configEditorRulesGet(w http.ResponseWriter, _ *http.Request) {
	if err := json.NewEncoder(w).Encode(config.Rules); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func configEditorRulesUpdate(w http.ResponseWriter, r *http.Request) {
	user, _, err := getAuthorizationFromRequest(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting authorized user").Error(), http.StatusInternalServerError)
		return
	}

	msg := &plugins.Rule{}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := patchConfig(cfg.Config, user, "", "Update rule", func(c *configFile) error {
		id := mux.Vars(r)["uuid"]

		for i := range c.Rules {
			if c.Rules[i].MatcherID() == id {
				c.Rules[i] = msg
			}
		}

		return nil
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
