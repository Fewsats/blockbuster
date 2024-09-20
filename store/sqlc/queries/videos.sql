-- name: CreateVideo :one
INSERT INTO videos (external_id, user_id, title, description, cover_url, price_in_cents, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetVideoByExternalID :one
SELECT * FROM videos
WHERE external_id = ? LIMIT 1;

-- name: ListUserVideos :many
SELECT * FROM videos
WHERE user_id = ?
ORDER BY created_at DESC;

-- name: DeleteVideo :exec
DELETE FROM videos
WHERE external_id = ?;

-- name: IncrementVideoViews :one
UPDATE videos
SET total_views = total_views + 1
WHERE external_id = ?
RETURNING *;

-- name: SearchVideos :many
SELECT * FROM videos
WHERE title LIKE ? OR description LIKE ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;


-- name: UpdateVideo :one
UPDATE videos
SET 
  thumbnail_url = COALESCE(sqlc.narg(thumbnail_url), thumbnail_url),
  duration_in_seconds = COALESCE(sqlc.narg(duration_in_seconds), duration_in_seconds),
  size_in_bytes = COALESCE(sqlc.narg(size_in_bytes), size_in_bytes),
  input_height = COALESCE(sqlc.narg(input_height), input_height),
  input_width = COALESCE(sqlc.narg(input_width), input_width),
  ready_to_stream = sqlc.arg(ready_to_stream)
WHERE external_id = sqlc.arg(external_id)
RETURNING *;