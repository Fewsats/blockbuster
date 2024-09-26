package cloudflare

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

type StreamsService struct {
	api       *cloudflare.API
	accountID string
}

func NewStreamsService(cfg *Config) (*StreamsService, error) {
	api, err := cloudflare.NewWithAPIToken(cfg.APIToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudflare API client: %w", err)
	}

	return &StreamsService{
		api: api,

		accountID: cfg.AccountID,
	}, nil
}

func (s *StreamsService) generateVideoUploadURL(ctx context.Context) (string, string, error) {

	expiry := time.Now().Add(2 * time.Hour)
	params := cloudflare.StreamCreateVideoParameters{
		AccountID:          s.accountID,
		MaxDurationSeconds: 3600, // 1 hour max duration
		Expiry:             &expiry,
		AllowedOrigins:     []string{"*"}, // TODO(pol) allow only our front-end
		RequireSignedURLs:  false,
	}

	// Call the API to create a direct upload URL
	result, err := s.api.StreamCreateVideoDirectURL(ctx, params)
	if err != nil {
		return "", "", fmt.Errorf("failed to create Stream direct upload URL: %w", err)
	}

	return result.UploadURL, result.UID, nil
}

func (s *StreamsService) getStreamVideoInfo(ctx context.Context,
	externalID string) (*cloudflare.StreamVideo, error) {

	params := cloudflare.StreamParameters{
		AccountID: s.accountID,
		VideoID:   externalID,
	}

	video, err := s.api.StreamGetVideo(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error getting stream video: %w", err)
	}

	return &video, nil
}

func (s *StreamsService) generateStreamURL(ctx context.Context,
	externalID string) (string, string, error) {

	expiry := time.Now().Add(23 * time.Hour).Unix()

	params := cloudflare.StreamSignedURLParameters{
		AccountID: s.accountID,
		VideoID:   externalID,
		EXP:       int(expiry),
	}

	token, err := s.api.StreamCreateSignedURL(ctx, params)
	if err != nil {
		return "", "", fmt.Errorf("error generating signed URL: %w", err)
	}

	HLSURL := fmt.Sprintf("https://customer-%s.cloudflarestream.com/%s/manifest/video.m3u8", s.accountID, token)
	DashURL := fmt.Sprintf("https://customer-%s.cloudflarestream.com/%s/manifest/video.mpd", s.accountID, token)

	return HLSURL, DashURL, nil
}
