package l402

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/fewsats/blockbuster/utils"
	"gopkg.in/macaroon.v2"
)

var (
	// byteOrder is the byte order used to encode/decode a macaroon's raw
	// identifier.
	byteOrder = binary.BigEndian

	// ErrMissingAuthorizationHeader is returned when the Authorization header is
	// missing.
	ErrMissingAuthorizationHeader = errors.New("missing Authorization header")

	// ErrMissingL402Header is returned when the L402 Authorization header is
	// missing.
	ErrMissingL402Header = errors.New("missing L402 Authorization header")

	// ErrInvalidPreimage is returned when the preimage is invalid.
	ErrInvalidPreimage = errors.New("invalid preimage")
)

// Authenticator is an authenticator that uses L402 tokens.
type Authenticator struct {
	provider InvoiceProvider
	clock    utils.Clock

	store  Store
	cfg    *Config
	logger *slog.Logger
}

// NewAuthenticator creates a new L402 authenticator.
func NewAuthenticator(logger *slog.Logger, provider InvoiceProvider,
	cfg *Config, store Store, clock utils.Clock) *Authenticator {

	return &Authenticator{
		provider: provider,
		clock:    clock,

		cfg:    cfg,
		store:  store,
		logger: logger,
	}
}

func (l *Authenticator) mintMacaroon(location string, pubKey, rootKey []byte,
	caveats map[string]string) (*macaroon.Macaroon, error) {

	mac, err := macaroon.New(rootKey, pubKey, location,
		macaroon.LatestVersion)
	if err != nil {
		return nil, fmt.Errorf("unable to create macaroon: %v", err)
	}

	for key, value := range caveats {
		if key == "expires_at" {
			// Validate the expires_at format
			_, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return nil, fmt.Errorf("invalid expires_at format: %v", err)
			}
		}

		rawCaveat := []byte(fmt.Sprintf("%s=%s", key, value))
		err := mac.AddFirstPartyCaveat(rawCaveat)
		if err != nil {
			return nil, fmt.Errorf("unable to add caveat(%s,%s): %v", key,
				value, err)
		}

	}

	return mac, nil
}

// NewL402Challenge creates a new L402 challenge (macaroon, invoice).
func (l *Authenticator) NewChallenge(ctx context.Context, productName string,
	pubKeyHex string, priceInUSDCents uint64,
	caveats map[string]string) (*Challenge, error) {

	// Convert pubKeyHex to [32]byte
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return nil, fmt.Errorf("unable to decode pubKeyHex: %v", err)
	}

	var pubKey [32]byte
	copy(pubKey[:], pubKeyBytes)

	// Create an invoice.
	lnInvocie, err := l.provider.CreateInvoice(
		ctx, priceInUSDCents, "USD", productName,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create invoice: %v", err)
	}

	// Create a random token ID and root key to identify the user/key pair.
	var randomRootKey [32]byte
	_, err = rand.Read(randomRootKey[:])
	if err != nil {
		return nil, fmt.Errorf("unable to generate random root key: %v", err)
	}

	// Store token ID and root key.
	err = l.store.CreateRootKey(ctx, pubKey, randomRootKey)
	if err != nil {
		return nil, fmt.Errorf("unable to store root key: %v", err)
	}

	paymentHash, err := hex.DecodeString(lnInvocie.PaymentHash)
	if err != nil {
		return nil, fmt.Errorf("unable to decode payment hash: %v", err)
	}

	var buf bytes.Buffer
	if err := binary.Write(&buf, byteOrder, uint16(0)); err != nil {
		return nil, fmt.Errorf("unable to write version: %v", err)
	}

	if _, err := buf.Write(paymentHash[:]); err != nil {
		return nil, fmt.Errorf("unable to write payment hash: %v", err)
	}

	if _, err := buf.Write(pubKey[:]); err != nil {
		return nil, fmt.Errorf("unable to write token ID: %v", err)
	}

	location := "fewsats.com"
	mac, err := l.mintMacaroon(location, buf.Bytes(), randomRootKey[:],
		caveats)
	if err != nil {
		return nil, fmt.Errorf("unable to create macaroon: %v", err)
	}

	return NewChallenge(mac, lnInvocie), nil
}

// ValidateL402Credentials validates the L402 credentials in the Authorization
// header.
func (l *Authenticator) ValidateL402Credentials(ctx context.Context,
	authHeader string) (string, error) {

	creds, err := l.ExtractCredentials(authHeader)
	if err != nil {
		return "", fmt.Errorf("unable to extract credentials: %w", err)
	}

	err = l.ValidateCredentials(ctx, creds)
	if err != nil {
		return "", fmt.Errorf("unable to validate credentials: %w", err)
	}

	payment_hash_hex := hex.EncodeToString(creds.PaymentHash[:])

	return payment_hash_hex, nil
}

// ExtractL402Credentials extracts the L402 credentials from the Authorization
// header.
func (l *Authenticator) ExtractCredentials(authHeader string) (*Credentials,
	error) {

	if authHeader == "" {
		return nil, ErrMissingAuthorizationHeader
	}

	// Make sure we support old macaroon prefix.
	authHeader = strings.Replace(authHeader, "LSAT", "L402", 1)

	// Check if the Authorization header has the L402 prefix.
	if !strings.HasPrefix(authHeader, "L402 ") {
		return nil, ErrMissingL402Header
	}

	// Extract the L402 token (macaroon and preimage).
	l402Token := strings.Split(strings.TrimPrefix(authHeader, "L402 "), ":")
	if len(l402Token) != 2 {
		return nil, fmt.Errorf("invalid L402 token: %s", authHeader)
	}

	// Extract the macaroon and preimage.
	// NOTE: They may provide multiple macaroon separated by comma, by now we
	// only support one macaroon.
	macBase64 := strings.Split(l402Token[0], ",")[0]
	preimageHex := l402Token[1]

	return DecodeL402Credentials(macBase64, preimageHex)
}

// ValidateL402Credentials validates the L402 credentials in the Authorization
// header.
//
// TODO(positiveblue): add req context to check the caveats.
func (l *Authenticator) ValidateCredentials(ctx context.Context,
	creds *Credentials) error {

	err := creds.ValidatePreimage()
	if err != nil {
		return ErrInvalidPreimage
	}

	// Retrieve the root key for the token ID.
	rootKey, err := l.store.GetRootKey(ctx, creds.TokenID)
	if err != nil {
		return fmt.Errorf("unable to retrieve root key: %v", err)
	}

	err = creds.VerifyMacaroon(rootKey)
	if err != nil {
		return fmt.Errorf("unable to verify macaroon: %v", err)
	}

	return nil
}

func (l *Authenticator) ValidateSignature(pubKeyHex, signatureHex,
	domain string, timestamp int64) error {

	// Check if the timestamp is within 10 minutes of the current time
	if time.Since(time.Unix(timestamp, 0)) > 10*time.Minute {
		return fmt.Errorf("timestamp is too old")
	}

	// TODO(pol) set up domain properly instead of hardcoding
	// Check if the domain is valid
	if domain != l.cfg.Domain {
		return fmt.Errorf("invalid domain")
	}

	// Create the message to be signed
	message := fmt.Sprintf("%s:%d", domain, timestamp)

	// Verify the signature
	err := verifySignature(pubKeyHex, signatureHex, message)
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}

	return nil
}

func verifySignature(pubKeyHex, signatureHex, message string) error {
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return fmt.Errorf("invalid public key hex: %w", err)
	}

	pubkey, err := schnorr.ParsePubKey(pubKeyBytes)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}

	sig, err := schnorr.ParseSignature(sigBytes)
	if err != nil {
		return fmt.Errorf("failed to parse signature: %w", err)
	}

	hash := sha256.Sum256([]byte(message))
	if !sig.Verify(hash[:], pubkey) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}
