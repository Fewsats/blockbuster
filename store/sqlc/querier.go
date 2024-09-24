// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package sqlc

import (
	"context"
	"time"
)

type Querier interface {
	CreateToken(ctx context.Context, arg CreateTokenParams) (Token, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (int64, error)
	CreateVideo(ctx context.Context, arg CreateVideoParams) (Video, error)
	DeleteExpiredTokens(ctx context.Context, expiration time.Time) error
	DeleteToken(ctx context.Context, token string) error
	DeleteVideo(ctx context.Context, externalID string) error
	GetOfferByPaymentHash(ctx context.Context, paymentHash string) (Offer, error)
	GetPurchaseByPaymentHash(ctx context.Context, paymentHash string) (Purchase, error)
	GetRootKeyByTokenID(ctx context.Context, tokenID []byte) ([]byte, error)
	GetToken(ctx context.Context, token string) (Token, error)
	GetUserByID(ctx context.Context, id int64) (GetUserByIDRow, error)
	GetUserIDByEmail(ctx context.Context, email string) (int64, error)
	GetVideoByExternalID(ctx context.Context, externalID string) (Video, error)
	IncrementVideoViews(ctx context.Context, externalID string) error
	InsertMacaroonToken(ctx context.Context, arg InsertMacaroonTokenParams) (int64, error)
	InsertOffer(ctx context.Context, arg InsertOfferParams) (int64, error)
	InsertPurchase(ctx context.Context, arg InsertPurchaseParams) (int64, error)
	ListUserVideos(ctx context.Context, userID int64) ([]Video, error)
	SearchVideos(ctx context.Context, arg SearchVideosParams) ([]Video, error)
	UpdateUserVerified(ctx context.Context, arg UpdateUserVerifiedParams) error
	UpdateVideo(ctx context.Context, arg UpdateVideoParams) (Video, error)
	VerifyToken(ctx context.Context, arg VerifyTokenParams) (string, error)
}

var _ Querier = (*Queries)(nil)
