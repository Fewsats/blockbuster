package auth

import (
	"context"
	"time"
)

type InvoiceStatus struct {
	PaymentHash string    `json:"payment_hash"`
	Preimage    string    `json:"preimage"`
	Settled     bool      `json:"settled"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type Store interface {
	GetUserByID(ctx context.Context, id int64) (User, error)
	GetOrCreateUserByEmail(ctx context.Context, email string) (int64, error)
	UpdateUserLightningAddress(ctx context.Context, id int64,
		lightningAddress string) error
	UpdateUserVerified(ctx context.Context, email string, verified bool) error
	StoreToken(ctx context.Context, email, token string,
		expiration time.Time) error
	VerifyToken(ctx context.Context, token string) (string, error)

	// GetInvoiceStatus retrieves the invoice status for a given payment hash.
	// Used to cache requests.
	GetInvoiceStatus(ctx context.Context,
		paymentHash string) (*InvoiceStatus, error)

	// UpsertInvoiceStatus upserts the invoice status for a given payment hash.
	// and returns the updated invoice status.
	UpsertInvoiceStatus(ctx context.Context, paymentHash, preimage string,
		settled bool) (*InvoiceStatus, error)
}
