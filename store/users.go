package store

import (
	"context"

	"github.com/fewsats/blockbuster/auth"
	"github.com/fewsats/blockbuster/store/sqlc"
)

// User methods
func (s *Store) CreateUser(ctx context.Context, email string) (int64, error) {
	userID, err := s.queries.CreateUser(ctx, email)

	return int64(userID), err
}

func (s *Store) GetUser(ctx context.Context, email string) (*auth.User, error) {
	user, err := s.queries.GetUser(ctx, email)
	if err != nil {
		return nil, err
	}

	return &auth.User{
		ID:       user.ID,
		Email:    user.Email,
		Verified: user.Verified,
	}, nil
}

func (s *Store) UpdateUserVerified(ctx context.Context, email string, verified bool) error {
	return s.queries.UpdateUserVerified(ctx, sqlc.UpdateUserVerifiedParams{
		Email:    email,
		Verified: verified,
	})
}
