-- name: InsertMacaroonToken :one
INSERT INTO macaroon_credentials (
    identifier, root_key, created_at, encoded_base_macaroon, disabled
) VALUES (
    ?, ?, ?, ?, ?
) RETURNING id;

-- name: GetRootKeyByIdentifier :one
SELECT root_key
FROM macaroon_credentials
WHERE identifier = ?;
