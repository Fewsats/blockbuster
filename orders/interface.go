package orders

import (
	"context"
	"time"
)

// Offer represents an offer that is available for purchase.
type Offer struct {
	// ID is the unique identifier of the offer.
	ID uint64 `json:"id"`

	// UserID is the user ID of the user selling this service.
	UserID uint64 `json:"user_id"`

	// ExternalID is the unique identifier of the service that this
	// offer is associated with.
	ExternalID string `json:"external_id"`

	// PaymentHash is the payment hash for the offer.
	PaymentHash string `json:"payment_hash"`

	// PriceInCents is the price of the item in cents.
	PriceInCents uint64 `json:"price_in_cents"`

	// Currency is the currency of the amount.
	Currency string `json:"currency"`

	// ExpirationDate is the expiration date for the credentials linked to
	// this offer. Only used in the Subscription-based pricing plans.
	ExpirationDate *time.Time `json:"expiration_date"`

	// CreatedAt is the timestamp when the offer was created.
	CreatedAt time.Time `json:"created_at"`
}

// Purchase represents a purchase made by the end clients.
type Purchase struct {
	// ID is the unique identifier of the purchase.
	ID uint64 `json:"id"`

	// UserID is the user ID of the user selling this service.
	UserID uint64 `json:"user_id"`

	// ExternalID is the external ID of the service.
	ExternalID string `json:"external_id"`

	// ServiceType is the type of the service. (STaaS, DaaS, FaaS...)
	ServiceType string `json:"service_type"`

	// PaymentHash is the payment hash for the offer linked to this
	// purchase.
	PaymentHash string `json:"payment_hash"`

	// PriceInCents is the price of the item in cents.
	PriceInCents uint64 `json:"price_in_cents"`

	// Currency is the currency used for the transaction.
	Currency string `json:"currency"`

	// ExpirationDate is the expiration date for the credentials linked to this
	// purchase.
	ExpirationDate *time.Time `json:"expiration_date"`

	// CreatedAt is the timestamp when the purchase was created.
	CreatedAt time.Time `json:"created_at"`
}

// Store is the interface for storing and retrieving order related data.
type Store interface {
	// InsertOffer inserts a new offer into the store.
	InsertOffer(ctx context.Context, offer *Offer) (uint64, error)

	// GetOfferByPaymentHash returns the offer for the given payment hash.
	GetOfferByPaymentHash(ctx context.Context, payreq string) (*Offer, error)

	// InsertPurchase inserts a new purchase into the store.
	InsertPurchase(ctx context.Context, purchase *Purchase) (uint64, error)

	// GetPurchaseByPaymentHash returns the purchase for the given payment hash.
	GetPurchaseByPaymentHash(ctx context.Context, payreq string) (*Purchase,
		error)
}

// NotificationService is the interface for sending notifications.
type NotificationService interface {
	// RegisterNewPurchaseEvent sends a notification for a new purchase.
	RegisterNewPurchaseEvent(externalID, paymentHash, email string) error
}
