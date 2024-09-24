package store

import (
	"context"
	"fmt"

	"github.com/fewsats/blockbuster/store/sqlc"
)

// CreateRootKey stores the root key for a given token ID.
func (s *Store) CreateRootKey(ctx context.Context, tokenID [32]byte,
	rootKey [32]byte) error {

	timestamp := s.clock.Now()
	txBody := func(queries *sqlc.Queries) error {
		params := sqlc.InsertMacaroonTokenParams{
			TokenID:   tokenID[:],
			RootKey:   rootKey[:],
			CreatedAt: timestamp,
		}

		_, err := queries.InsertMacaroonToken(ctx, params)
		if err != nil {
			return err
		}

		return nil
	}

	if err := s.ExecTx(ctx, txBody); err != nil {
		return fmt.Errorf("failed to insert new root key metadata: %v", err)
	}

	return nil
}

// GetRootKey retrieves the root key for a given token ID.
func (s *Store) GetRootKey(ctx context.Context, tokenID [32]byte) ([32]byte,
	error) {

	var rootKey [32]byte
	txBody := func(queries *sqlc.Queries) error {
		row, err := queries.GetRootKeyByTokenID(ctx, tokenID[:])
		if err != nil {
			return err
		}

		copy(rootKey[:], row)

		return nil
	}

	if err := s.ExecTx(ctx, txBody); err != nil {
		return rootKey, fmt.Errorf("failed to get root key metadata: %v", err)
	}

	return rootKey, nil
}
