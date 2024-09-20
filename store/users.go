package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fewsats/blockbuster/auth"
	"github.com/fewsats/blockbuster/store/sqlc"
	_ "github.com/mattn/go-sqlite3"
)

// User methods
func (s *Store) CreateUser(ctx context.Context, email string) (int64, error) {
	userID, err := s.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:     email,
		CreatedAt: s.clock.Now(),
	})
	return int64(userID), err
}

func (s *Store) GetUserIDByEmail(ctx context.Context, email string) (uint64,
	error) {

	userID, err := s.queries.GetUserIDByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, auth.ErrUserNotFound
		}

		return 0, fmt.Errorf("failed to get user ID by email: %w", err)
	}

	return uint64(userID), nil
}

func (s *Store) UpdateUserVerified(ctx context.Context,
	email string, verified bool) error {

	return s.queries.UpdateUserVerified(ctx, sqlc.UpdateUserVerifiedParams{
		Email:    email,
		Verified: verified,
	})
}

func (s *Store) GetOrCreateUserByEmail(ctx context.Context,
	email string) (int64, error) {

	userID, err := s.queries.GetUserIDByEmail(ctx, email)

	if err != nil && !errors.Is(err, auth.ErrUserNotFound) {
		return s.queries.CreateUser(ctx, sqlc.CreateUserParams{
			Email:     email,
			CreatedAt: s.clock.Now(),
		})
	}

	return userID, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (auth.User, error) {
	user, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.User{}, auth.ErrUserNotFound
		}

		return auth.User{}, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return auth.User{
		ID:       user.ID,
		Email:    user.Email,
		Verified: user.Verified,
	}, nil
}
