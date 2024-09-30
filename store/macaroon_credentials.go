package store

import (
	"context"
	"fmt"

	"github.com/fewsats/blockbuster/store/sqlc"
)

// CreateRootKey stores the root key for a given token ID.
func (s *Store) CreateRootKey(ctx context.Context, identifier string,
	rootKey string, encodedBaseMacaroon string) error {

	timestamp := s.clock.Now()
	txBody := func(queries *sqlc.Queries) error {
		params := sqlc.InsertMacaroonTokenParams{
			Identifier:          identifier,
			RootKey:             rootKey,
			CreatedAt:           timestamp,
			EncodedBaseMacaroon: encodedBaseMacaroon,
			Disabled:            false,
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
func (s *Store) GetRootKey(ctx context.Context, identifier string) (string,
	error) {

	var rootKey string
	txBody := func(queries *sqlc.Queries) error {
		row, err := queries.GetRootKeyByIdentifier(ctx, identifier)
		if err != nil {
			return err
		}
		rootKey = row
		return nil
	}

	if err := s.ExecTx(ctx, txBody); err != nil {
		return "", fmt.Errorf("failed to get root key: %v", err)
	}

	return rootKey, nil
}
