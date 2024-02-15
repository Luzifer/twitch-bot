package kofi

import "time"

type (
	hookPayload struct {
		VerificationToken          string       `json:"verification_token"`
		MessageID                  string       `json:"message_id"`
		Timestamp                  time.Time    `json:"timestamp"`
		Type                       hookType     `json:"type"`
		IsPublic                   bool         `json:"is_public"`
		FromName                   string       `json:"from_name"`
		Message                    *string      `json:"message"`
		Amount                     float64      `json:"amount,string"`
		URL                        string       `json:"url"`
		Email                      string       `json:"email"`
		Currency                   string       `json:"currency"`
		IsSubscriptionPayment      bool         `json:"is_subscription_payment"`
		IsFirstSubscriptionPayment bool         `json:"is_first_subscription_payment"`
		KofiTransactionID          string       `json:"kofi_transaction_id"`
		ShopItems                  []shopItem   `json:"shop_items"`
		TierName                   *string      `json:"tier_name"`
		Shipping                   shippingInfo `json:"shipping"`
	}

	hookType string

	shippingInfo struct {
		FullName        string `json:"full_name"`
		StreetAddress   string `json:"street_address"`
		City            string `json:"city"`
		StateOrProvince string `json:"state_or_province"`
		PostalCode      string `json:"postal_code"`
		Country         string `json:"country"`
		CountryCode     string `json:"country_code"`
		Telephone       string `json:"telephone"`
	}

	shopItem struct {
		DirectLinkCode string `json:"direct_link_code"`
		VariationName  string `json:"variation_name"`
		Quantity       int    `json:"quantity"`
	}
)

const (
	hookTypeCommission   hookType = "Commission"
	hookTypeDonation     hookType = "Donation"
	hookTypeShopOrder    hookType = "Shop Order"
	hookTypeSubscription hookType = "Subscription"
)
