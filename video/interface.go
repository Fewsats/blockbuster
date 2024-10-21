package video

import (
	"context"
	"io"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/fewsats/blockbuster/l402"
)

type Authenticator interface {
	ValidateSignature(pubKeyHex, signatureHex, domain string,
		timestamp int64) error

	NewChallenge(ctx context.Context, domain, pubKeyHex string,
		priceInCents uint64, caveats map[string]string) (*l402.Challenge, error)

	ValidateL402Credentials(ctx context.Context,
		authHeader string) (string, error)
}

type Store interface {
	GetOrCreateUserByEmail(ctx context.Context, email string) (int64, error)

	CreateVideo(ctx context.Context, params CreateVideoParams) (*Video, error)
	UpdateCloudflareInfo(ctx context.Context, externalID string,
		params *CloudflareVideoInfo) (*Video, error)

	GetVideoByExternalID(ctx context.Context, externalID string) (*Video, error)
	ListUserVideos(ctx context.Context, userID int64) ([]*Video, error)

	// IncrementVideoViews increments the views of a video by 1.
	IncrementVideoViews(ctx context.Context, externalID string) error

	UpdateVideoInfo(ctx context.Context, externalID string,
		params *UpdateVideoInfoParams) (*Video, error)
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
	CreateOffer(ctx context.Context, userID int64,
		PriceInCents uint64, externalID, paymentHash string) error

	// RecordPurchase creates a new purchase if there is not one already for
	// the given payment hash.
	RecordPurchase(ctx context.Context, paymentHash, serviceType string) error
}

type CloudflareService interface {
	GetStreamVideoInfo(ctx context.Context,
		externalID string) (*cloudflare.StreamVideo, error)
	GenerateVideoUploadURL(ctx context.Context) (string, string, error)
	UploadPublicFile(ctx context.Context, key, prefix string,
		reader io.ReadSeeker) (string, error)
	GenerateStreamURL(ctx context.Context, externalID string) (string, string, error)
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

type CloudflareVideoInfo struct {
	ThumbnailURL      string
	DurationInSeconds float64
	SizeInBytes       int64
	InputHeight       int32
	InputWidth        int32
	ReadyToStream     bool
	HLSURL            string
	DashURL           string
}

type Video struct {
	ID          int64  `json:"-"`
	ExternalID  string `json:"external_id"`
	UserID      int64  `json:"-"`
	L402URL     string `json:"l402_url"`
	L402InfoURI string `json:"l402_info_uri"`

	Title          string `json:"title"`
	Description    string `json:"description"`
	CoverURL       string `json:"cover_url"`
	PriceInCents   int64  `json:"price_in_cents"`
	TotalViews     int64  `json:"total_views"`
	TotalPurchases int64  `json:"total_purchases"`

	ThumbnailURL      string  `json:"-"`
	HlsURL            string  `json:"-"`
	DashURL           string  `json:"-"`
	DurationInSeconds float64 `json:"duration_in_seconds"`
	SizeInBytes       int64   `json:"size_in_bytes"`
	InputHeight       int32   `json:"input_height"`
	InputWidth        int32   `json:"input_width"`

	ReadyToStream bool      `json:"ready_to_stream"`
	CreatedAt     time.Time `json:"created_at"`
}

type UpdateVideoInfoParams struct {
	Title        string
	Description  string
	PriceInCents int64
}
