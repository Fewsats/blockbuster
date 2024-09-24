package cloudflare

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

const (
	// ExpirationTime is the time after which a file download link expires.
	ExpirationTime = 24 * time.Hour * 30 // 30 days
)

// Service is the main storage service interface.
type Service struct {
	r2      *R2Service
	streams *StreamsService

	cfg *Config
}

// NewService creates a new storage service.
func NewService(cfg *Config) (*Service, error) {
	r2, err := NewR2Service(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create R2 service: %w", err)
	}

	streams, err := NewStreamsService(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Streams service: %w", err)
	}

	return &Service{
		r2:      r2,
		streams: streams,

		cfg: cfg,
	}, nil
}

// PublicFileURL returns the URL of a file in the storage provider.
func (s *Service) PublicFileURL(key string) string {
	return s.r2.publicFileURL(key)
}

// GenerateVideoViewURL generates a presigned URL for a video in the storage provider.
func (s *Service) GenerateStreamURL(ctx context.Context,
	key string) (string, error) {

	return s.streams.generateStreamURL(ctx, key)
}

func (s *Service) GetStreamVideoInfo(ctx context.Context,
	externalID string) (*cloudflare.StreamVideo, error) {

	return s.streams.getStreamVideoInfo(ctx, externalID)
}

func (s *Service) GenerateVideoUploadURL(ctx context.Context) (string,
	string, error) {

	return s.streams.generateVideoUploadURL(ctx)
}

func (s *Service) UploadPublicFile(ctx context.Context, fileID string,
	prefix string, reader io.ReadSeeker) (string, error) {

	key := fmt.Sprintf("%s/%s", prefix, fileID)
	return s.r2.uploadPublicFile(ctx, key, reader)
}

func (s *Service) DeletePublicFile(ctx context.Context, key string) error {
	err := s.r2.deletePublicFile(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}
	return nil
}
