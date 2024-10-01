-- name: UpsertInvoiceStatus :one
INSERT INTO invoice_status (payment_hash, settled, preimage, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT (payment_hash) DO UPDATE SET
    settled = EXCLUDED.settled,
    preimage = EXCLUDED.preimage,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: GetInvoiceStatus :one
SELECT * FROM invoice_status WHERE payment_hash = ?;