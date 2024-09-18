-- name: CreateUser :one
INSERT INTO users (email, verified)
VALUES (?, false)
RETURNING id;

-- name: GetUser :one
SELECT * FROM users
WHERE email = ? LIMIT 1;

-- name: UpdateUserVerified :exec
UPDATE users
SET verified = ?
WHERE email = ?