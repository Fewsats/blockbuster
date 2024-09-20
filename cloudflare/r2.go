package cloudflare

import (
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type R2Service struct {
	r2           *s3.S3
	videoBucket  string
	publicBucket string
}

func NewR2Service(cfg *Config) (*R2Service, error) {
	cred := credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretAccessKey, "")

	r2Config := &aws.Config{
		Credentials: cred,
		Endpoint:    cfg.Endpoint,
		// Endpoint:         aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)),
		Region:           aws.String("auto"),
		S3ForcePathStyle: aws.Bool(true),
	}

	sess, err := session.NewSession(r2Config)
	if err != nil {
		return nil, err
	}

	r2 := s3.New(sess)

	return &R2Service{
		r2: r2,

		videoBucket:  cfg.VideoBucketName,
		publicBucket: cfg.PublicBucketName,
	}, nil
}

const (
	defaultExpireTime = 1 * time.Hour
)

func (r *R2Service) VideoURL(key string) string {
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", r.videoBucket, key)
}

func (r *R2Service) PublicFileURL(key string) string {
	// TODO(pol) this is a dev access to staging bucket hardcoded
	return fmt.Sprintf("https://pub-3c55410f5c574362bbaa52948499969e.r2.dev/%s", key)
	// return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", r.publicBucket, key)
}

func (r *R2Service) GenerateVideoViewURL(key string) (string, error) {
	req, _ := r.r2.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(r.videoBucket),
		Key:    aws.String(key),
	})
	return req.Presign(defaultExpireTime)
}

// GenerateVideoUploadURL generates a presigned URL for a video in the storage provider.
func (r *R2Service) GenerateVideoUploadURL(key string) (string, error) {
	req, _ := r.r2.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(r.videoBucket),
		Key:    aws.String(key),
	})
	return req.Presign(defaultExpireTime)
}

func (r *R2Service) DeleteVideoFile(key string) error {
	_, err := r.r2.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(r.videoBucket),
		Key:    aws.String(key),
	})
	return err
}

func (r *R2Service) UploadPublicFile(key string, reader io.ReadSeeker) (string, error) {
	_, err := r.r2.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(r.publicBucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		return "", err
	}
	return r.PublicFileURL(key), nil
}

func (r *R2Service) DeletePublicFile(key string) error {
	_, err := r.r2.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(r.publicBucket),
		Key:    aws.String(key),
	})
	return err
}
