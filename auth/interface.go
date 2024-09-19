package auth

import (
	"context"
	"time"
)

type Store interface {
	StoreToken(ctx context.Context, email, token string, expiration time.Time) error
	VerifyToken(ctx context.Context, token string) (string, error)
}
