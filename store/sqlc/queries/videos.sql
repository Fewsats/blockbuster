-- name: CreateVideo :one
INSERT INTO videos (external_id, user_id, title, description, cover_url, price_in_cents, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetVideoByExternalID :one
SELECT v.*, COUNT(p.id) as total_purchases
FROM videos v
LEFT JOIN purchases p ON v.external_id = p.external_id
WHERE v.external_id = ?
GROUP BY v.id
LIMIT 1;

-- name: ListUserVideos :many
SELECT v.*, COUNT(p.id) as total_purchases
FROM videos v
LEFT JOIN purchases p ON v.external_id = p.external_id
WHERE v.user_id = ?
GROUP BY v.id
ORDER BY v.created_at DESC;

-- name: DeleteVideo :exec
DELETE FROM videos
WHERE external_id = ?;

-- name: IncrementVideoViews :exec
UPDATE videos
SET total_views = total_views + 1
WHERE external_id = ?;

-- name: SearchVideos :many
SELECT * FROM videos
WHERE title LIKE ? OR description LIKE ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;


-- name: UpdateCloudflareInfo :one
UPDATE videos
SET 
  thumbnail_url = COALESCE(sqlc.narg(thumbnail_url), thumbnail_url),
  hls_url = COALESCE(sqlc.narg(hls_url), hls_url),
  dash_url = COALESCE(sqlc.narg(dash_url), dash_url),
  duration_in_seconds = COALESCE(sqlc.narg(duration_in_seconds), duration_in_seconds),
  size_in_bytes = COALESCE(sqlc.narg(size_in_bytes), size_in_bytes),
  input_height = COALESCE(sqlc.narg(input_height), input_height),
  input_width = COALESCE(sqlc.narg(input_width), input_width),
  ready_to_stream = sqlc.arg(ready_to_stream)
WHERE external_id = sqlc.arg(external_id)
RETURNING *;

-- name: UpdateVideoInfo :one
UPDATE videos
SET 
  title = COALESCE(sqlc.narg(title), title),
  description = COALESCE(sqlc.narg(description), description),
  price_in_cents = COALESCE(sqlc.narg(price_in_cents), price_in_cents)
WHERE external_id = sqlc.arg(external_id)
RETURNING *;