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

func registerEditorAutoMessageRoutes() {
	for _, rd := range []plugins.HTTPRouteRegistrationArgs{
		{
			Description:         "Returns the current set of configured auto-messages in JSON format",
			HandlerFunc:         configEditorHandleAutoMessagesGet,
			Method:              http.MethodGet,
			Module:              moduleConfigEditor,
			Name:                "Get current auto-messages",
			Path:                "/auto-messages",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeJSON,
		},
		{
			Description:         "Adds a new Auto-Message",
			HandlerFunc:         configEditorHandleAutoMessageAdd,
			Method:              http.MethodPost,
			Module:              moduleConfigEditor,
			Name:                "Add Auto-Message",
			Path:                "/auto-messages",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeTextPlain,
		},
		{
			Description:         "Deletes the given Auto-Message",
			HandlerFunc:         configEditorHandleAutoMessageDelete,
			Method:              http.MethodDelete,
			Module:              moduleConfigEditor,
			Name:                "Delete Auto-Message",
			Path:                "/auto-messages/{uuid}",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeTextPlain,
			RouteParams: []plugins.HTTPRouteParamDocumentation{
				{
					Description: "UUID of the auto-message to delete",
					Name:        "uuid",
					Required:    false,
					Type:        "string",
				},
			},
		},
		{
			Description:         "Updates the given Auto-Message",
			HandlerFunc:         configEditorHandleAutoMessageUpdate,
			Method:              http.MethodPut,
			Module:              moduleConfigEditor,
			Name:                "Update Auto-Message",
			Path:                "/auto-messages/{uuid}",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeTextPlain,
			RouteParams: []plugins.HTTPRouteParamDocumentation{
				{
					Description: "UUID of the auto-message to update",
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

func configEditorHandleAutoMessageAdd(w http.ResponseWriter, r *http.Request) {
	user, _, err := getAuthorizationFromRequest(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting authorized user").Error(), http.StatusInternalServerError)
		return
	}

	msg := &autoMessage{}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg.UUID = uuid.Must(uuid.NewV4()).String()

	if err := patchConfig(cfg.Config, user, "", "Add auto-message", func(c *configFile) error {
		c.AutoMessages = append(c.AutoMessages, msg)
		return nil
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func configEditorHandleAutoMessageDelete(w http.ResponseWriter, r *http.Request) {
	user, _, err := getAuthorizationFromRequest(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting authorized user").Error(), http.StatusInternalServerError)
		return
	}

	if err := patchConfig(cfg.Config, user, "", "Delete auto-message", func(c *configFile) error {
		var (
			id  = mux.Vars(r)["uuid"]
			tmp []*autoMessage
		)

		for i := range c.AutoMessages {
			if c.AutoMessages[i].ID() == id {
				continue
			}
			tmp = append(tmp, c.AutoMessages[i])
		}

		c.AutoMessages = tmp

		return nil
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func configEditorHandleAutoMessagesGet(w http.ResponseWriter, _ *http.Request) {
	if err := json.NewEncoder(w).Encode(config.AutoMessages); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func configEditorHandleAutoMessageUpdate(w http.ResponseWriter, r *http.Request) {
	user, _, err := getAuthorizationFromRequest(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting authorized user").Error(), http.StatusInternalServerError)
		return
	}

	msg := &autoMessage{}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := patchConfig(cfg.Config, user, "", "Update auto-message", func(c *configFile) error {
		id := mux.Vars(r)["uuid"]

		for i := range c.AutoMessages {
			if c.AutoMessages[i].ID() == id {
				c.AutoMessages[i] = msg
			}
		}

		return nil
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
