// Package kofi contains a webhook listener to be used in the Ko-fi
// API to receive information about (recurring) donations / shop orders
package kofi

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "kofi"

type (
	kofiDonationEvent struct {
		Amount            float64 `field:"amount" description:"Amount donated as submitted by Ko-fi"`
		Channel           string  `field:"channel" description:"Channel the event occurred for"`
		Currency          string  `field:"currency" description:"Currency of the donated amount"`
		From              string  `field:"from" description:"Name submitted by the donor"`
		IsFirstSubPayment bool    `field:"isFirstSubPayment" description:"Whether this is the first subscription payment"`
		IsSubscription    bool    `field:"isSubscription" description:"Whether this is a subscription payment"`
		Message           *string `field:"message,omitzero" description:"Message entered by the donor"`
		Tier              *string `field:"tier,omitzero" description:"Subscription tier"`
	}
)

var (
	eventCreatorFunc plugins.TypedEventHandlerFunc
	getModuleConfig  plugins.ModuleConfigGetterFunc

	ptrStringEmpty = func(s string) *string { return &s }("")
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) (err error) {
	eventCreatorFunc = args.CreateTypedEvent
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

// Event implements event.Event
func (kofiDonationEvent) Event() *string { return new("kofi_donation") }

func handleKoFiPost(w http.ResponseWriter, r *http.Request) {
	channel := mux.Vars(r)["channel"]

	channelModuleConf := getModuleConfig(actorName, channel)

	// The data is sent (posted) with a content type of application/x-www-form-urlencoded.
	// A field named 'data' contains the payment information as a JSON string.
	jsonData := r.FormValue("data") //#nosec:G120 // Request body size is limited by API route registration middleware
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

	switch payload.Type {
	case hookTypeDonation, hookTypeSubscription:
		// Single or Recurring Donation
		evt := kofiDonationEvent{
			Amount:            payload.Amount,
			Channel:           "#" + strings.TrimLeft(channel, "#"),
			Currency:          payload.Currency,
			From:              payload.FromName,
			IsFirstSubPayment: payload.IsFirstSubscriptionPayment,
			IsSubscription:    payload.IsSubscriptionPayment,
		}

		if payload.IsPublic && payload.Message != nil {
			evt.Message = payload.Message
		}

		if payload.IsSubscriptionPayment && payload.TierName != nil {
			evt.Tier = payload.TierName
		}

		if err = eventCreatorFunc(evt); err != nil {
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
