package store

import (
	"context"
	"database/sql"

	"github.com/fewsats/blockbuster/auth"
	"github.com/fewsats/blockbuster/store/sqlc"
)

func (s *Store) GetInvoiceStatus(ctx context.Context,
	paymentHash string) (*auth.InvoiceStatus, error) {

	status, err := s.queries.GetInvoiceStatus(ctx, paymentHash)
	if err != nil {
		return nil, err
	}
	return &auth.InvoiceStatus{
		PaymentHash: status.PaymentHash,
		Settled:     status.Settled,
		UpdatedAt:   status.UpdatedAt,
		CreatedAt:   status.CreatedAt,
		Preimage:    status.Preimage.String,
	}, nil
}

func (s *Store) UpsertInvoiceStatus(ctx context.Context, paymentHash,
	preimage string, settled bool) (*auth.InvoiceStatus, error) {

	timestamp := s.clock.Now()

	status, err := s.queries.UpsertInvoiceStatus(ctx, sqlc.UpsertInvoiceStatusParams{
		PaymentHash: paymentHash,
		Settled:     settled,
		Preimage:    sql.NullString{String: preimage, Valid: preimage != ""},
		UpdatedAt:   timestamp,
		CreatedAt:   timestamp,
	})
	if err != nil {
		return nil, err
	}

	return &auth.InvoiceStatus{
		PaymentHash: status.PaymentHash,
		Settled:     status.Settled,
		Preimage:    status.Preimage.String,
		UpdatedAt:   status.UpdatedAt,
		CreatedAt:   status.CreatedAt,
	}, nil
}
