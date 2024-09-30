package l402

import (
	"context"

	"github.com/fewsats/blockbuster/lightning"
)

// InvoiceProvider is the interface for creating new LN invoices.
type InvoiceProvider interface {
	// CreateInvoice creates a new LN invoice for the given price and
	// description.
	CreateInvoice(ctx context.Context, amount uint64, currency string,
		description string) (*lightning.LNInvoice, error)
}

type Store interface {
	// CreateRootKey stores the root key for a given token ID.
	CreateRootKey(ctx context.Context, identifier string, rootKey string,
		encodedBaseMacaroon string) error

	// GetRootKey retrieves the root key for a given token ID.
	GetRootKey(ctx context.Context, identifier string) (string, error)
}
