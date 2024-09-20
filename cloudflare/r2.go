package cloudflare

import (
	"context"
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
		Credentials:      cred,
		Endpoint:         cfg.Endpoint,
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

func (r *R2Service) videoURL(key string) string {
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", r.videoBucket, key)
}

func (r *R2Service) publicFileURL(key string) string {
	// TODO(pol) this is a dev access to staging bucket hardcoded
	return fmt.Sprintf("https://pub-3c55410f5c574362bbaa52948499969e.r2.dev/%s", key)
	// return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", r.publicBucket, key)
}

func (r *R2Service) uploadPublicFile(ctx context.Context, key string, reader io.ReadSeeker) (string, error) {
	_, err := r.r2.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(r.publicBucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		return "", err
	}
	return r.publicFileURL(key), nil
}

func (r *R2Service) deletePublicFile(ctx context.Context, key string) error {
	_, err := r.r2.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(r.publicBucket),
		Key:    aws.String(key),
	})
	return err
}
