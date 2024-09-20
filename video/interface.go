package video

import (
	"context"
	"io"
	"time"
)

type Store interface {
	GetOrCreateUserByEmail(ctx context.Context, email string) (int64, error)
	CreateVideo(ctx context.Context, params CreateVideoParams) (*Video, error)
	GetVideoByExternalID(ctx context.Context, externalID string) (*Video, error)
	ListUserVideos(ctx context.Context, userID int64) ([]*Video, error)
	DeleteVideo(ctx context.Context, externalID string) error
	// SearchVideos(ctx context.Context, query string, limit, offset int32) ([]*Video, error)
}

// StorageProvider is the interface for interacting with a storage provider.
type StorageProvider interface {
	// FileURL returns the URL of a file in the storage provider.
	// FileURL(key string) string

	// GenerateUploadURL generates a presigned URL for uploading a
	// file directly to Cloudflare R2.
	GenerateVidoUploadURL(key string) (string, error)
}

// PublicStorage is the interface for managing public files
type PublicStorage interface {
	// UploadPublicFile uploads a public file to the storage provider.
	UploadPublicFile(fileID, prefix string, reader io.ReadSeeker) (string, error)
}

type CreateVideoParams struct {
	ExternalID   string
	UserID       int64
	Title        string
	Description  string
	VideoURL     string
	CoverURL     string
	PriceInCents int64
}

type Video struct {
	ID           int64     `json:"-"`
	ExternalID   string    `json:"external_id"`
	UserID       int64     `json:"-"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	CoverURL     string    `json:"cover_url"`
	VideoURL     string    `json:"video_url"`
	PriceInCents int64     `json:"price_in_cents"`
	TotalViews   int64     `json:"total_views"`
	CreatedAt    time.Time `json:"created_at"`
}
