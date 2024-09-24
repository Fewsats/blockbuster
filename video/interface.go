package video

import (
	"context"
	"io"
	"time"
)

type Store interface {
	GetOrCreateUserByEmail(ctx context.Context, email string) (int64, error)
	CreateVideo(ctx context.Context, params CreateVideoParams) (*Video, error)
	UpdateVideo(ctx context.Context, externalID string,
		params UpdateVideoParams) (*Video, error)

	GetVideoByExternalID(ctx context.Context, externalID string) (*Video, error)
	ListUserVideos(ctx context.Context, userID int64) ([]*Video, error)

	// IncrementVideoViews increments the views of a video by 1.
	IncrementVideoViews(ctx context.Context, externalID string) error
}

// NotificationService is the interface for sending notifications.
type NotificationService interface {
	// RegisterNewPurchaseEvent sends a notification for a new purchase.
	RegisterNewPurchaseEvent(externalID, paymentHash, email string) error

	// RegisterNewVideoUploadEvent sends a notification for a new video upload.
	RegisterNewVideoUploadEvent(externalID, email string) error
}

type OrdersMgr interface {
	// CreateOffer creates a new offer.
	CreateOffer(ctx context.Context, externalID, payHash string) error

	// RecordPurchase creates a new purchase if there is not one already for
	// the given payment hash.
	RecordPurchase(ctx context.Context, payHash, serviceType string) error
}

type CloudflareService interface {
	GetStreamVideoInfo(ctx context.Context, videoID string) (*UpdateVideoParams, error)
	GenerateVideoUploadURL(ctx context.Context) (string, string, error)
	UploadPublicFile(key, prefix string, reader io.ReadSeeker) (string, error)
	GenerateStreamURL(ctx context.Context, videoID string) (string, error)
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

type UpdateVideoParams struct {
	ThumbnailURL      string
	DurationInSeconds float64
	SizeInBytes       int64
	InputHeight       int32
	InputWidth        int32
	ReadyToStream     bool
}

type Video struct {
	ID         int64  `json:"-"`
	ExternalID string `json:"external_id"`
	UserID     int64  `json:"-"`

	Title        string `json:"title"`
	Description  string `json:"description"`
	CoverURL     string `json:"cover_url"`
	PriceInCents int64  `json:"price_in_cents"`
	TotalViews   int64  `json:"total_views"`

	ThumbnailURL      string  `json:"thumbnail_url"`
	DurationInSeconds float64 `json:"duration_in_seconds"`
	SizeInBytes       int64   `json:"size_in_bytes"`
	InputHeight       int32   `json:"input_height"`
	InputWidth        int32   `json:"input_width"`

	ReadyToStream bool      `json:"ready_to_stream"`
	CreatedAt     time.Time `json:"created_at"`
}
