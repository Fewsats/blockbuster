package lightning

// Amount represents an amount in a specific currency.
type Amount struct {
	// Amount is the amount in the currency's smallest unit.
	Amount uint64 `json:"amount"`

	// Currency is the currency of the amount.
	Currency string `json:"currency"`
}

// LNInvoice represents a Lightning Network invoice.
type LNInvoice struct {
	// UserAmount is the amount of the invoice in the user's currency.
	UserAmount Amount

	// PaymentAmount is the amount of the invoice in the payment currency.
	PaymentAmount Amount

	// PaymentHash is the hash of the payment preimage.
	PaymentHash string

	// PaymentRequest is the Lightning Network invoice payment request.
	PaymentRequest string
}
