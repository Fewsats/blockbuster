package cloudflare

import (
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
	r2 *R2Service

	cfg *Config
}

// NewService creates a new storage service.
func NewService(cfg *Config) (*Service, error) {
	r2, err := NewR2Service(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create R2 service: %w", err)
	}

	return &Service{
		r2:  r2,
		cfg: cfg,
	}, nil
}

// // FileURL returns the URL of a file in the storage provider.
// func (s *Service) PublicFileURL(key string) string {
// 	return s.r2.PublicFileURL(key)
// }

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

func (s *Service) GenerateVideoUploadURL(key string) (string, error) {
	return s.r2.GenerateVideoUploadURL(key)
}

func (s *Service) UploadPublicFile(fileID string, prefix string, reader io.ReadSeeker) (string, error) {

	key := fmt.Sprintf("%s/%s", prefix, fileID)
	return s.r2.UploadPublicFile(key, reader)
}
