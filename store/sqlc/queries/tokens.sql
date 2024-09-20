-- name: CreateToken :one
INSERT INTO tokens (email, token, expiration, created_at)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetToken :one
SELECT * FROM tokens
WHERE token = ? LIMIT 1;

-- name: DeleteToken :exec
DELETE FROM tokens
WHERE token = ?;

-- name: DeleteExpiredTokens :exec
DELETE FROM tokens
WHERE expiration < ?;

-- name: VerifyToken :one
SELECT email FROM tokens
WHERE token = ? AND expiration > ?
LIMIT 1;