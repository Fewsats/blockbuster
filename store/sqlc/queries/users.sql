-- name: CreateUser :one
INSERT INTO users (email, verified, created_at)
VALUES (?, false, ?)
RETURNING id;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = ? LIMIT 1;

-- name: GetUserIDByEmail :one
SELECT id FROM users
WHERE email = ? LIMIT 1;

-- name: UpdateUserVerified :exec
UPDATE users
SET verified = ?
WHERE email = ?;

-- name: UpdateUserLightningAddress :exec
UPDATE users
SET lightning_address = ?
WHERE id = ?;

