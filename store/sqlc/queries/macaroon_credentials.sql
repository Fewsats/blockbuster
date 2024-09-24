-- name: InsertMacaroonToken :one
INSERT INTO macaroon_credentials (
    token_id, root_key, created_at
) VALUES (
    ?, ?, ?
) RETURNING id;

-- name: GetRootKeyByTokenID :one
SELECT root_key
FROM macaroon_credentials
WHERE token_id = ?;
