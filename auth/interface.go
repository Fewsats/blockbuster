package auth

import (
	"context"
	"time"
)

type Store interface {
	GetUserByID(ctx context.Context, id int64) (User, error)
	GetOrCreateUserByEmail(ctx context.Context, email string) (int64, error)
	UpdateUserLightningAddress(ctx context.Context, id int64,
		lightningAddress string) error
	UpdateUserVerified(ctx context.Context, email string, verified bool) error
	StoreToken(ctx context.Context, email, token string,
		expiration time.Time) error
	VerifyToken(ctx context.Context, token string) (string, error)
}
