-- name: CreateVideo :one
INSERT INTO videos (external_id, user_id, title, description, video_url, cover_url, price_in_cents)
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