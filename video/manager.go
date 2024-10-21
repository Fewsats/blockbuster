package video

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"time"

	"github.com/fewsats/blockbuster/l402"
	"github.com/fewsats/blockbuster/utils"
)

// Manager is the main video service interface.
type Manager struct {
	orders        OrdersMgr
	cf            CloudflareService
	authenticator Authenticator

	store  Store
	clock  utils.Clock
	logger *slog.Logger
}

// NewManager creates a new storage service.
func NewManager(orders OrdersMgr, cf CloudflareService,
	authenticator Authenticator, store Store, logger *slog.Logger,
	clock utils.Clock) *Manager {

	return &Manager{
		authenticator: authenticator,
		cf:            cf,
		orders:        orders,

		clock:  clock,
		store:  store,
		logger: logger,
	}
}

// updateVideoIfNeeded retrieves the video info from cloudflare and updates the video in the database
// if the video is not ready to stream.
// When uploaded videos are set as not ready to stream by default.
func (m *Manager) IsVideoReady(ctx context.Context, externalID string) error {
	video, err := m.store.GetVideoByExternalID(ctx, externalID)
	if err != nil {
		m.logger.Error("Failed to get video by ID", "error", err)
		return fmt.Errorf("failed to fetch video: %w", err)
	}
	// first time the video is accessed we'll populate it with the cloudflare info
	if !video.ReadyToStream {
		videoInfo, err := m.cf.GetStreamVideoInfo(ctx, externalID)
		if err != nil {
			m.logger.Error("Failed to get video info", "error", err)
			return fmt.Errorf("failed to get video info: %w", err)
		}

		video, err = m.store.UpdateCloudflareInfo(ctx, externalID, &CloudflareVideoInfo{
			ThumbnailURL:      videoInfo.Thumbnail,
			DashURL:           videoInfo.Playback.Dash,
			HLSURL:            videoInfo.Playback.HLS,
			DurationInSeconds: videoInfo.Duration,
			SizeInBytes:       int64(videoInfo.Size),
			InputHeight:       int32(videoInfo.Input.Height),
			InputWidth:        int32(videoInfo.Input.Width),
			ReadyToStream:     videoInfo.ReadyToStream,
		})
		if err != nil {
			m.logger.Error("Failed to update video", "error", err)
			return fmt.Errorf("failed to update video: %w", err)
		}
	}

	// We already attempted to update the info, if video ain't ready here, it aint ready.
	if !video.ReadyToStream {
		return fmt.Errorf("video is not ready to stream: %w", err)
	}

	return nil
}

func (m *Manager) GenerateStreamURL(ctx context.Context,
	externalID string) (string, string, error) {

	HLSURL, DashURL, err := m.cf.GenerateStreamURL(ctx, externalID)
	if err != nil {
		m.logger.Error("Failed to generate presigned URL", "error", err)
	}

	return HLSURL, DashURL, nil
}

// CreateChallenge creates a new L402 challenge for downloading a file from our
// storage service.
func (m *Manager) CreateL402Challenge(ctx context.Context, pubKeyHex string,
	externalID string) (*l402.Challenge, error) {

	video, err := m.store.GetVideoByExternalID(ctx, externalID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video: %w", err)
	}

	expiresAt := m.clock.Now().Add(ExpirationTime)
	caveats := map[string]string{
		"external_id": video.ExternalID,
		"expires_at":  expiresAt.Format(time.RFC3339),
	}

	creds, err := m.authenticator.NewChallenge(ctx, video.Title, pubKeyHex,
		uint64(video.PriceInCents), caveats)
	if err != nil {
		return nil, fmt.Errorf("failed to create L402 challenge: %w", err)
	}

	paymentHash := creds.Invoice.PaymentHash
	err = m.orders.CreateOffer(ctx, video.UserID, uint64(video.PriceInCents),
		video.ExternalID, paymentHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create offer for file(%s): %w",
			video.ExternalID, err)
	}

	return creds, nil
}

func (m *Manager) RecordPurchaseAndView(ctx context.Context, externalID, paymentHash,
	serviceType string) error {

	err := m.orders.RecordPurchase(ctx, paymentHash, serviceType)
	if err != nil {
		return fmt.Errorf("failed to record purchase: %w", err)
	}

	err = m.store.IncrementVideoViews(ctx, externalID)
	if err != nil {
		return fmt.Errorf("failed to increment video views: %w", err)
	}

	return nil
}

func (m *Manager) ProcessAndUploadCoverImage(gCtx context.Context,
	externalID string, coverImageHeader *multipart.FileHeader) (string, error) {

	file, err := coverImageHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open cover image: %w", err)
	}
	defer file.Close()

	coverImageBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read cover image: %w", err)
	}

	coverImageReader := bytes.NewReader(coverImageBytes)

	coverURL, err := m.cf.UploadPublicFile(gCtx, externalID,
		"cover-images", coverImageReader)
	if err != nil {
		return "", fmt.Errorf("failed to upload cover file: %w", err)
	}

	return coverURL, nil
}

func (m *Manager) GenerateVideoUploadURL(ctx context.Context) (string, string, error) {
	uploadURL, externalID, err := m.cf.GenerateVideoUploadURL(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate upload URL: %w", err)
	}
	return uploadURL, externalID, nil
}

func (m *Manager) PrepareVideoUpload(ctx context.Context, userID int64,
	req UploadVideoRequest) (string, string, error) {

	uploadURL, externalID, err := m.GenerateVideoUploadURL(ctx)
	if err != nil {
		m.logger.Error("Failed to generate upload URL", "error", err)
		return "", "", fmt.Errorf("failed to generate upload URL: %w", err)
	}

	coverURL, err := m.ProcessAndUploadCoverImage(ctx, externalID, req.CoverImage)
	if err != nil {
		m.logger.Error("Failed to upload cover image", "error", err)
		return "", "", fmt.Errorf("failed to upload cover image: %w", err)
	}

	_, err = m.store.CreateVideo(ctx, CreateVideoParams{
		ExternalID:   externalID,
		UserID:       userID,
		Title:        req.Title,
		Description:  req.Description,
		CoverURL:     coverURL,
		PriceInCents: req.PriceInCents,
	})

	if err != nil {
		m.logger.Error("Failed to create video", "error", err)
		return "", "", fmt.Errorf("failed to save video metadata: %w", err)
	}

	return uploadURL, externalID, nil
}

func (m *Manager) UpdateVideoInfo(ctx context.Context, externalID string, req UpdateVideoInfoRequest) (*Video, error) {
	video, err := m.store.UpdateVideoInfo(ctx, externalID, &UpdateVideoInfoParams{
		Title:        req.Title,
		Description:  req.Description,
		PriceInCents: req.PriceInCents,
	})
	if err != nil {
		m.logger.Error("Failed to update video info", "error", err)
		return nil, fmt.Errorf("failed to update video info: %w", err)
	}
	return video, nil
}

func (m *Manager) DeleteVideo(ctx context.Context, externalID string) error {
	// Then, delete the video from the database
	err := m.store.DeleteVideo(ctx, externalID)
	if err != nil {
		return fmt.Errorf("failed to delete video from database: %w", err)
	}

	return nil
}
