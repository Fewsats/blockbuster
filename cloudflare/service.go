package cloudflare

import (
	"context"
	"fmt"
	"io"
	"time"
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
	return s.r2.PublicFileURL(key)
}

// VideoURL returns the URL of a video in the storage provider.
func (s *Service) VideoURL(key string) string {
	return s.r2.VideoURL(key)
}

// GenerateVideoViewURL generates a presigned URL for a video in the storage provider.
func (s *Service) GenerateVideoViewURL(key string) (string, error) {
	return s.r2.GenerateVideoViewURL(key)
}

func (s *Service) DeletePublicFile(key string) error {
	err := s.r2.DeletePublicFile(key)
	if err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}
	return nil
}

func (s *Service) DeleteVideoFile(key string) error {
	err := s.r2.DeleteVideoFile(key)
	if err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}
	return nil
}

func (s *Service) GenerateVideoUploadURL(ctx context.Context) (string, string, error) {
	return s.streams.GenerateVideoUploadURL(ctx)
}

func (s *Service) UploadPublicFile(fileID string,
	prefix string, reader io.ReadSeeker) (string, error) {

	key := fmt.Sprintf("%s/%s", prefix, fileID)
	return s.r2.UploadPublicFile(key, reader)
}
