package store

import (
	"context"
	"fmt"
	"time"

	"github.com/fewsats/blockbuster/store/sqlc"
)

// Token methods
func (s *Store) GetToken(ctx context.Context, token string) (*sqlc.Token, error) {
	t, err := s.queries.GetToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Add the new StoreToken method
func (s *Store) StoreToken(ctx context.Context, email, token string, expiration time.Time) error {
	_, err := s.queries.CreateToken(ctx, sqlc.CreateTokenParams{
		Token:      token,
		Email:      email,
		Expiration: expiration,
	})

	return err
}

func (s *Store) ExecTx(ctx context.Context, txBody func(*sqlc.Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	q := s.queries.WithTx(tx)
	err = txBody(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

func (s *Store) VerifyToken(ctx context.Context, token string) (string, error) {
	params := sqlc.VerifyTokenParams{
		Token:      token,
		Expiration: s.clock.Now(),
	}
	email, err := s.queries.VerifyToken(ctx, params)
	if err != nil {
		return "", err
	}
	return email, nil
}
