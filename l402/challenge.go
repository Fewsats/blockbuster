package l402

import (
	"encoding/base64"
	"fmt"

	"github.com/fewsats/blockbuster/lightning"
	"gopkg.in/macaroon.v2"
)

const (
	// ChallengeHeaderValueFormat is the format for the L402 challenge header value.
	ChallengeHeaderValueFormat = "L402 macaroon=\"%s\", invoice=\"%s\""
)

// Challenge represents an L402 challenge.
//
// NOTE: an L402 challenge has two components:
// - Credentials
// - Payment request
// In the current version of the L402 protocol (V0), the credentials are a
// macaroon and the payment request is a Lightning Network invoice.
type Challenge struct {
	// Macaroon is the credentials for the L402 challenge in V0.
	Macaroon *macaroon.Macaroon

	// Invoice is the Lightning invoice used as payment request for the L402
	// challenge in V0.
	Invoice *lightning.LNInvoice
}

// NewChallenge creates a new L402 challenge.
func NewChallenge(macaroon *macaroon.Macaroon,
	invoice *lightning.LNInvoice) *Challenge {

	return &Challenge{
		Macaroon: macaroon,
		Invoice:  invoice,
	}
}

// EncodedCredentials returns the encoded credentials for the L402 challenge.
func (c *Challenge) EncodedCredentials() (string, error) {
	mac, err := c.Macaroon.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("unable to marshal macaroon: %v", err)
	}

	return base64.StdEncoding.EncodeToString(mac), nil
}

// EncodedPaymentRequest returns the encoded payment request for the L402
// challenge.
func (c *Challenge) EncodedPaymentRequest() (string, error) {
	if c.Invoice.PaymentRequest == "" {
		return "", fmt.Errorf("payment request is empty")
	}

	return c.Invoice.PaymentRequest, nil
}

// HeaderKey returns the header key for the L402 challenge.
func (c *Challenge) HeaderKey() string {
	return "WWW-Authenticate"
}

// HeaderValue returns the header value for the L402 challenge.
func (c *Challenge) HeaderValue() (string, error) {
	creds, err := c.EncodedCredentials()
	if err != nil {
		return "", err
	}

	invoice, err := c.EncodedPaymentRequest()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(ChallengeHeaderValueFormat, creds, invoice), nil
}
