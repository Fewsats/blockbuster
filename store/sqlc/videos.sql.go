// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: videos.sql

package sqlc

import (
	"context"
	"database/sql"
)

const createVideo = `-- name: CreateVideo :one
INSERT INTO videos (user_email, title, description, file_path, thumbnail_path, price)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id, user_email, title, description, file_path, thumbnail_path, price, total_views, created_at
`

type CreateVideoParams struct {
	UserEmail     string
	Title         string
	Description   sql.NullString
	FilePath      string
	ThumbnailPath sql.NullString
	Price         float64
}

func (q *Queries) CreateVideo(ctx context.Context, arg CreateVideoParams) (Video, error) {
	row := q.db.QueryRowContext(ctx, createVideo,
		arg.UserEmail,
		arg.Title,
		arg.Description,
		arg.FilePath,
		arg.ThumbnailPath,
		arg.Price,
	)
	var i Video
	err := row.Scan(
		&i.ID,
		&i.UserEmail,
		&i.Title,
		&i.Description,
		&i.FilePath,
		&i.ThumbnailPath,
		&i.Price,
		&i.TotalViews,
		&i.CreatedAt,
	)
	return i, err
}

const deleteVideo = `-- name: DeleteVideo :exec
DELETE FROM videos
WHERE id = ?
`

func (q *Queries) DeleteVideo(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteVideo, id)
	return err
}

const getVideo = `-- name: GetVideo :one
SELECT id, user_email, title, description, file_path, thumbnail_path, price, total_views, created_at FROM videos
WHERE id = ? LIMIT 1
`

func (q *Queries) GetVideo(ctx context.Context, id int64) (Video, error) {
	row := q.db.QueryRowContext(ctx, getVideo, id)
	var i Video
	err := row.Scan(
		&i.ID,
		&i.UserEmail,
		&i.Title,
		&i.Description,
		&i.FilePath,
		&i.ThumbnailPath,
		&i.Price,
		&i.TotalViews,
		&i.CreatedAt,
	)
	return i, err
}

const incrementVideoViews = `-- name: IncrementVideoViews :one
UPDATE videos
SET total_views = total_views + 1
WHERE id = ?
RETURNING id, user_email, title, description, file_path, thumbnail_path, price, total_views, created_at
`

func (q *Queries) IncrementVideoViews(ctx context.Context, id int64) (Video, error) {
	row := q.db.QueryRowContext(ctx, incrementVideoViews, id)
	var i Video
	err := row.Scan(
		&i.ID,
		&i.UserEmail,
		&i.Title,
		&i.Description,
		&i.FilePath,
		&i.ThumbnailPath,
		&i.Price,
		&i.TotalViews,
		&i.CreatedAt,
	)
	return i, err
}

const listUserVideos = `-- name: ListUserVideos :many
SELECT id, user_email, title, description, file_path, thumbnail_path, price, total_views, created_at FROM videos
WHERE user_email = ?
ORDER BY created_at DESC
`

func (q *Queries) ListUserVideos(ctx context.Context, userEmail string) ([]Video, error) {
	rows, err := q.db.QueryContext(ctx, listUserVideos, userEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Video
	for rows.Next() {
		var i Video
		if err := rows.Scan(
			&i.ID,
			&i.UserEmail,
			&i.Title,
			&i.Description,
			&i.FilePath,
			&i.ThumbnailPath,
			&i.Price,
			&i.TotalViews,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listVideos = `-- name: ListVideos :many
SELECT id, user_email, title, description, file_path, thumbnail_path, price, total_views, created_at FROM videos
ORDER BY created_at DESC
LIMIT ? OFFSET ?
`

type ListVideosParams struct {
	Limit  int64
	Offset int64
}

func (q *Queries) ListVideos(ctx context.Context, arg ListVideosParams) ([]Video, error) {
	rows, err := q.db.QueryContext(ctx, listVideos, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Video
	for rows.Next() {
		var i Video
		if err := rows.Scan(
			&i.ID,
			&i.UserEmail,
			&i.Title,
			&i.Description,
			&i.FilePath,
			&i.ThumbnailPath,
			&i.Price,
			&i.TotalViews,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const searchVideos = `-- name: SearchVideos :many
SELECT id, user_email, title, description, file_path, thumbnail_path, price, total_views, created_at FROM videos
WHERE title LIKE ? OR description LIKE ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?
`

type SearchVideosParams struct {
	Title       string
	Description sql.NullString
	Limit       int64
	Offset      int64
}

func (q *Queries) SearchVideos(ctx context.Context, arg SearchVideosParams) ([]Video, error) {
	rows, err := q.db.QueryContext(ctx, searchVideos,
		arg.Title,
		arg.Description,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Video
	for rows.Next() {
		var i Video
		if err := rows.Scan(
			&i.ID,
			&i.UserEmail,
			&i.Title,
			&i.Description,
			&i.FilePath,
			&i.ThumbnailPath,
			&i.Price,
			&i.TotalViews,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateVideo = `-- name: UpdateVideo :one
UPDATE videos
SET title = ?, description = ?, thumbnail_path = ?, price = ?
WHERE id = ?
RETURNING id, user_email, title, description, file_path, thumbnail_path, price, total_views, created_at
`

type UpdateVideoParams struct {
	Title         string
	Description   sql.NullString
	ThumbnailPath sql.NullString
	Price         float64
	ID            int64
}

func (q *Queries) UpdateVideo(ctx context.Context, arg UpdateVideoParams) (Video, error) {
	row := q.db.QueryRowContext(ctx, updateVideo,
		arg.Title,
		arg.Description,
		arg.ThumbnailPath,
		arg.Price,
		arg.ID,
	)
	var i Video
	err := row.Scan(
		&i.ID,
		&i.UserEmail,
		&i.Title,
		&i.Description,
		&i.FilePath,
		&i.ThumbnailPath,
		&i.Price,
		&i.TotalViews,
		&i.CreatedAt,
	)
	return i, err
}
