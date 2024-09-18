-- name: CreateVideo :one
INSERT INTO videos (user_email, title, description, file_path, thumbnail_path, price)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetVideo :one
SELECT * FROM videos
WHERE id = ? LIMIT 1;

-- name: ListVideos :many
SELECT * FROM videos
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListUserVideos :many
SELECT * FROM videos
WHERE user_email = ?
ORDER BY created_at DESC;

-- name: UpdateVideo :one
UPDATE videos
SET title = ?, description = ?, thumbnail_path = ?, price = ?
WHERE id = ?
RETURNING *;

-- name: DeleteVideo :exec
DELETE FROM videos
WHERE id = ?;

-- name: IncrementVideoViews :one
UPDATE videos
SET total_views = total_views + 1
WHERE id = ?
RETURNING *;

-- name: SearchVideos :many
SELECT * FROM videos
WHERE title LIKE ? OR description LIKE ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;