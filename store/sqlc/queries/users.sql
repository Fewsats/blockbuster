-- name: CreateUser :one
INSERT INTO users (email, verified)
VALUES (?, false)
RETURNING id;

-- name: GetUserByID :one
SELECT id, email, verified FROM users
WHERE id = ? LIMIT 1;

-- name: GetUserIDByEmail :one
SELECT id FROM users
WHERE email = ? LIMIT 1;

-- name: UpdateUserVerified :exec
UPDATE users
SET verified = ?
WHERE email = ?

