// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: orders.sql

package sqlc

import (
	"context"
	"database/sql"
	"time"
)

const getOfferByPaymentHash = `-- name: GetOfferByPaymentHash :one
SELECT id, user_id, external_id, payment_hash, price_in_cents, currency, expiration_date, created_at
FROM offers
WHERE payment_hash = ?
`

func (q *Queries) GetOfferByPaymentHash(ctx context.Context, paymentHash string) (Offer, error) {
	row := q.db.QueryRowContext(ctx, getOfferByPaymentHash, paymentHash)
	var i Offer
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ExternalID,
		&i.PaymentHash,
		&i.PriceInCents,
		&i.Currency,
		&i.ExpirationDate,
		&i.CreatedAt,
	)
	return i, err
}

const getPurchaseByPaymentHash = `-- name: GetPurchaseByPaymentHash :one
SELECT id, user_id, external_id, service_type, price_in_cents, currency, expiration_date, payment_hash, created_at
FROM purchases
WHERE payment_hash = ?
`

func (q *Queries) GetPurchaseByPaymentHash(ctx context.Context, paymentHash string) (Purchase, error) {
	row := q.db.QueryRowContext(ctx, getPurchaseByPaymentHash, paymentHash)
	var i Purchase
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ExternalID,
		&i.ServiceType,
		&i.PriceInCents,
		&i.Currency,
		&i.ExpirationDate,
		&i.PaymentHash,
		&i.CreatedAt,
	)
	return i, err
}

const insertOffer = `-- name: InsertOffer :one
INSERT INTO offers (
    user_id, external_id, payment_hash, price_in_cents, currency, expiration_date,
    created_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
) RETURNING id
`

type InsertOfferParams struct {
	UserID         int64
	ExternalID     string
	PaymentHash    string
	PriceInCents   int64
	Currency       string
	ExpirationDate sql.NullTime
	CreatedAt      time.Time
}

func (q *Queries) InsertOffer(ctx context.Context, arg InsertOfferParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, insertOffer,
		arg.UserID,
		arg.ExternalID,
		arg.PaymentHash,
		arg.PriceInCents,
		arg.Currency,
		arg.ExpirationDate,
		arg.CreatedAt,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const insertPurchase = `-- name: InsertPurchase :one
INSERT INTO purchases (
    user_id, external_id, service_type, price_in_cents, currency,
    expiration_date, payment_hash, created_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING id
`

type InsertPurchaseParams struct {
	UserID         int64
	ExternalID     string
	ServiceType    string
	PriceInCents   int64
	Currency       string
	ExpirationDate sql.NullTime
	PaymentHash    string
	CreatedAt      time.Time
}

func (q *Queries) InsertPurchase(ctx context.Context, arg InsertPurchaseParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, insertPurchase,
		arg.UserID,
		arg.ExternalID,
		arg.ServiceType,
		arg.PriceInCents,
		arg.Currency,
		arg.ExpirationDate,
		arg.PaymentHash,
		arg.CreatedAt,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}
