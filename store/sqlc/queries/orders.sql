-- name: InsertOffer :one
INSERT INTO offers (
    user_id, external_id, payment_hash, price_in_cents, currency, expiration_date,
    created_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
) RETURNING id;

-- name: GetOfferByPaymentHash :one
SELECT *
FROM offers
WHERE payment_hash = ?;


-- name: InsertPurchase :one
INSERT INTO purchases (
    user_id, external_id, service_type, price_in_cents, currency,
    expiration_date, payment_hash, created_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING id;

-- name: GetPurchaseByPaymentHash :one
SELECT *
FROM purchases
WHERE payment_hash = ?;
