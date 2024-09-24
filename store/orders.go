package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/fewsats/blockbuster/orders"
	"github.com/fewsats/blockbuster/store/sqlc"
)

// InsertOffer inserts a new offer into the store.
func (s *Store) InsertOffer(ctx context.Context, offer *orders.Offer) (uint64, error) {
	if offer.ID != 0 {
		return 0, fmt.Errorf("trying to insert an offer with an ID: %d", offer.ID)
	}

	timestamp := s.clock.Now()
	var expirationDate time.Time
	if offer.ExpirationDate != nil {
		expirationDate = *offer.ExpirationDate
	}

	params := sqlc.InsertOfferParams{
		UserID:       int64(offer.UserID),
		ExternalID:   offer.ExternalID,
		PaymentHash:  offer.PaymentHash,
		PriceInCents: int64(offer.PriceInCents),
		Currency:     offer.Currency,
		CreatedAt:    timestamp,
		ExpirationDate: func(t *time.Time) sql.NullTime {
			if t == nil {
				return sql.NullTime{}
			}
			return sql.NullTime{Time: *t, Valid: true}
		}(&expirationDate),
	}

	var id uint64
	txBody := func(queries *sqlc.Queries) error {
		newID, err := queries.InsertOffer(ctx, params)
		if err != nil {
			return err
		}

		id = uint64(newID)
		offer.ID = uint64(newID)

		return nil
	}

	if err := s.ExecTx(ctx, txBody); err != nil {
		return 0, fmt.Errorf("failed to insert offer: %v", err)
	}

	return id, nil
}

// GetOfferByPaymentHash returns the offer for the given payment hash.
func (s *Store) GetOfferByPaymentHash(ctx context.Context,
	payreq string) (*orders.Offer, error) {

	var offer *orders.Offer
	txBody := func(queries *sqlc.Queries) error {
		row, err := queries.GetOfferByPaymentHash(ctx, payreq)
		if err != nil {
			return err
		}

		offer = &orders.Offer{
			ID:           uint64(row.ID),
			UserID:       uint64(row.UserID),
			ExternalID:   row.ExternalID,
			PaymentHash:  row.PaymentHash,
			PriceInCents: uint64(row.PriceInCents),
			Currency:     row.Currency,

			ExpirationDate: func(nt sql.NullTime) *time.Time {
				if nt.Valid {
					return &nt.Time
				}
				return nil
			}(row.ExpirationDate),
			CreatedAt: row.CreatedAt,
		}

		return nil
	}

	if err := s.ExecTx(ctx, txBody); err != nil {
		return nil, fmt.Errorf("failed to get offer by payment hash(%s): %v",
			payreq, err)
	}

	return offer, nil
}

// InsertPurchase inserts a new purchase into the store.
func (s *Store) InsertPurchase(ctx context.Context, purchase *orders.Purchase) (uint64, error) {
	if purchase.ID != 0 {
		return 0, fmt.Errorf("trying to insert a purchase with an ID: %d",
			purchase.ID)
	}

	timestamp := s.clock.Now()

	params := sqlc.InsertPurchaseParams{
		UserID:       int64(purchase.UserID),
		ExternalID:   purchase.ExternalID,
		ServiceType:  purchase.ServiceType,
		PaymentHash:  purchase.PaymentHash,
		PriceInCents: int64(purchase.PriceInCents),
		Currency:     purchase.Currency,
		ExpirationDate: func(t *time.Time) sql.NullTime {
			if t == nil {
				return sql.NullTime{}
			}
			return sql.NullTime{Time: *t, Valid: true}
		}(purchase.ExpirationDate),
		CreatedAt: timestamp,
	}

	var id uint64
	txBody := func(queries *sqlc.Queries) error {
		newID, err := queries.InsertPurchase(ctx, params)
		if err != nil {
			return err
		}

		id = uint64(newID)
		purchase.ID = uint64(newID)

		return nil
	}

	if err := s.ExecTx(ctx, txBody); err != nil {
		return 0, fmt.Errorf("failed to insert purchase: %v", err)
	}

	return id, nil
}

// GetPurchaseByPaymentHash returns the purchase for the given payment hash.
func (s *Store) GetPurchaseByPaymentHash(ctx context.Context,
	payreq string) (*orders.Purchase, error) {

	var purchase *orders.Purchase
	txBody := func(queries *sqlc.Queries) error {
		row, err := queries.GetPurchaseByPaymentHash(ctx, payreq)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return orders.ErrNotFound
			}
			return err
		}

		purchase = &orders.Purchase{
			ID:           uint64(row.ID),
			UserID:       uint64(row.UserID),
			ExternalID:   row.ExternalID,
			ServiceType:  row.ServiceType,
			PaymentHash:  row.PaymentHash,
			PriceInCents: uint64(row.PriceInCents),
			Currency:     row.Currency,
		}

		return nil
	}

	err := s.ExecTx(ctx, txBody)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase by payment hash(%s): %w",
			payreq, err)
	}

	return purchase, nil
}
