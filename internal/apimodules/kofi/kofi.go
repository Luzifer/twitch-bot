// Package kofi contains a webhook listener to be used in the Ko-fi
// API to receive information about (recurring) donations / shop orders
package kofi

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Luzifer/go_helpers/v2/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/plugins"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const actorName = "kofi"

var (
	eventCreatorFunc plugins.EventHandlerFunc
	getModuleConfig  plugins.ModuleConfigGetterFunc

	ptrStringEmpty = func(s string) *string { return &s }("")
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) (err error) {
	eventCreatorFunc = args.CreateEvent
	getModuleConfig = args.GetModuleConfigForChannel

	if err = args.RegisterAPIRoute(plugins.HTTPRouteRegistrationArgs{
		Description:  "Endpoint to handle Ko-fi Webhook posts",
		HandlerFunc:  handleKoFiPost,
		Method:       http.MethodPost,
		Module:       actorName,
		Name:         "Handle Ko-fi Webhook",
		Path:         "/webhook/{channel}",
		ResponseType: plugins.HTTPRouteResponseTypeJSON,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Channel to create the event in",
				Name:        "channel",
			},
		},
	}); err != nil {
		return fmt.Errorf("registering API route: %w", err)
	}

	return nil
}

func handleKoFiPost(w http.ResponseWriter, r *http.Request) {
	channel := mux.Vars(r)["channel"]

	channelModuleConf := getModuleConfig(actorName, channel)

	// The data is sent (posted) with a content type of application/x-www-form-urlencoded.
	// A field named 'data' contains the payment information as a JSON string.
	jsonData := r.FormValue("data")
	if jsonData == "" {
		// Well, no.
		logrus.WithField("remote_addr", r.RemoteAddr).Warn("received KoFi hook without payload")
		http.Error(w, "you missed something", http.StatusBadRequest)
		return
	}

	var (
		err     error
		payload hookPayload
	)

	// Read the payload
	if err = json.Unmarshal([]byte(jsonData), &payload); err != nil {
		logrus.WithError(err).Error("unmarshalling KoFi JSON data")
		http.Error(w, "that's not valid json, you know", http.StatusBadRequest)
		return
	}

	// If we know the verification token, validate the payload
	if validateToken := channelModuleConf.MustString("verification_token", ptrStringEmpty); validateToken != "" && payload.VerificationToken != validateToken {
		logrus.WithFields(logrus.Fields{
			"expected": fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(validateToken))),
			"provided": fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(payload.VerificationToken))),
		}).Error("received Ko-fi payload with invalid verification token")

		http.Error(w, "ehm, who are you?", http.StatusForbidden)
		return
	}

	fields := fieldcollection.NewFieldCollection()
	fields.Set("channel", "#"+strings.TrimLeft(channel, "#"))

	switch payload.Type {
	case hookTypeDonation, hookTypeSubscription:
		// Single or Recurring Donation
		fields.Set("from", payload.FromName)
		fields.Set("amount", payload.Amount)
		fields.Set("currency", payload.Currency)
		fields.Set("isSubscription", payload.IsSubscriptionPayment)
		fields.Set("isFirstSubPayment", payload.IsFirstSubscriptionPayment)

		if payload.IsPublic && payload.Message != nil {
			fields.Set("message", *payload.Message)
		}

		if payload.IsSubscriptionPayment && payload.TierName != nil {
			fields.Set("tier", *payload.TierName)
		}

		if err = eventCreatorFunc("kofi_donation", fields); err != nil {
			logrus.WithError(err).Error("creating kofi_donation event")
			http.Error(w, "ehm, that didn't work, I'm sorry", http.StatusInternalServerError)
			return
		}

	default:
		// Unsupported, we take that and discard it
		logrus.WithField("type", payload.Type).Warn("received unhandled hook type")
	}

	w.WriteHeader(http.StatusOK)
}
