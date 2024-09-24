package orders

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

var (
	// ErrNotFound is the error returned when the requested item is not found
	// in the store.
	ErrNotFound = errors.New("not found")
)

// Manager represents the manager for orders related operations.
type Manager struct {
	// store is the store for orders related data.
	store Store

	// logger is the logger for orders related operations.
	logger *slog.Logger
}

// NewManager creates a new orders manager.
func NewManager(logger *slog.Logger, store Store) *Manager {

	return &Manager{
		store:  store,
		logger: logger,
	}
}

// PurchaseInfo represents the information of a purchase for reporting purposes.
type PurchaseInfo struct {
	// CoverPicture is the cover picture of the purchase.
	CoverPicture string `json:"cover_picture"`

	// Title is the title of the purchase.
	Title string `json:"title"`

	// ExternalID is the external ID service of the purchase.
	ExternalID string `json:"external_id"`

	// PaymentHash is the payment hash of the purchase.
	PaymentHash string `json:"payment_hash"`

	// CreatedAt is the creation date of the purchase.
	CreatedAt string `json:"created_at"`

	// Amount is the amount of the purchase.
	Amount uint64 `json:"amount"`

	// Currency is the currency of the purchase.
	Currency string `json:"currency"`
}

func (m *Manager) CreateOffer(ctx context.Context, userID int64,
	PriceInCents uint64, externalID, paymentHash string) error {

	offer := &Offer{
		UserID:       uint64(userID),
		ExternalID:   externalID,
		PaymentHash:  paymentHash,
		PriceInCents: PriceInCents,
		Currency:     "USD",

		ExpirationDate: nil,
	}

	_, err := m.store.InsertOffer(ctx, offer)
	if err != nil {
		return fmt.Errorf("failed to insert offer for %s: %w", externalID, err)
	}

	return nil
}

// RecordPurchase creates a new purchase if there is not one already for
// the given payment hash.
func (m *Manager) RecordPurchase(ctx context.Context, payHash, serviceType string) error {
	_, err := m.store.GetPurchaseByPaymentHash(ctx, payHash)
	if err == nil {
		// Purchase already exists, nothing to record.
		return nil
	}

	if !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("failed to get purchase by payment hash(%s): %w",
			payHash, err)
	}

	offer, err := m.store.GetOfferByPaymentHash(ctx, payHash)
	if err != nil {
		return fmt.Errorf("failed to get offer by payment hash(%s): %w",
			payHash, err)
	}

	// TODO(pol) check this fields are saved to DB and populated correctly
	purchase := &Purchase{
		UserID:       offer.UserID,
		ExternalID:   offer.ExternalID,
		ServiceType:  serviceType,
		PaymentHash:  payHash,
		PriceInCents: offer.PriceInCents,
		Currency:     offer.Currency,

		ExpirationDate: offer.ExpirationDate,
	}

	_, err = m.store.InsertPurchase(ctx, purchase)
	if err != nil {
		return fmt.Errorf("failed to insert purchase for offer %d: %w",
			offer.ID, err)
	}

	m.logger.Info("Purchase recorded",
		"externalID", purchase.ExternalID,
		"userID", purchase.UserID,
		"name", purchase.ExternalID,
		"serviceType", purchase.ServiceType)

	return nil
}
