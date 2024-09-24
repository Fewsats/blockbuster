package lightning

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var (
	// AlbySupportedCurrencies is the list of currencies that we support for
	// creating invoices using Alby's API.
	AlbySupportedCurrencies = []string{"BTC", "USD"}
)

// AlbyInvoiceProvider is an implementation of the InvoiceProvider interface
// that uses Alby's API to create new LN invoices.
type AlbyInvoiceProvider struct {
	Client HTTPClient
	APIKey string
}

// NewAlbyProvider creates a new InvoiceProvider with the given API key.
func NewAlbyProvider(client HTTPClient, apiKey string) *AlbyInvoiceProvider {
	return &AlbyInvoiceProvider{
		Client: client,
		APIKey: apiKey,
	}
}

// AlbyInvoiceData represents the data required to create a new LN invoice using
// Alby's API.
type AlbyInvoiceData struct {
	Amount      uint64 `json:"amount"`
	Description string `json:"description"`
	Currency    string `json:"currency"`
}

// AlbyInvoiceResponse represents the response from Alby's API when creating a
// new LN invoice.
type AlbyInvoiceResponse struct {
	Amount         uint64 `json:"amount"`
	RHashStr       string `json:"r_hash_str"`
	PaymentRequest string `json:"payment_request"`
	ExpiresAt      string `json:"expires_at"`
}

// supportedCurrency returns true if the given currency is supported by Alby's
// API.
func (a *AlbyInvoiceProvider) supportedCurrency(currency string) bool {
	for _, c := range AlbySupportedCurrencies {
		if c == currency {
			return true
		}
	}

	return false
}

// CreateInvoice creates a new LN invoice for the given price and
// description. It returns the payment request and the payment hash
// hex-encoded.
func (a *AlbyInvoiceProvider) CreateInvoice(ctx context.Context, amount uint64,
	currency string, description string) (*LNInvoice, error) {

	if !a.supportedCurrency(currency) {
		return nil, fmt.Errorf("currency %s not supported", currency)
	}

	data := AlbyInvoiceData{
		Amount:      amount,
		Description: description,
		Currency:    currency,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost
	url := "https://api.getalby.com/invoices"
	reqBody := bytes.NewBuffer(jsonData)

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+a.APIKey)
	req.Header.Set("Content-Type", "application/json")


	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to create invoice, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var invoiceResponse AlbyInvoiceResponse
	err = json.Unmarshal(body, &invoiceResponse)
	if err != nil {
		return nil, err
	}

	return &LNInvoice{
		UserAmount:     Amount{Amount: amount, Currency: currency},
		PaymentAmount:  Amount{Amount: invoiceResponse.Amount, Currency: "BTC"},
		PaymentHash:    invoiceResponse.RHashStr,
		PaymentRequest: invoiceResponse.PaymentRequest,
	}, nil
}
