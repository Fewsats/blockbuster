package l402

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/fewsats/blockbuster/lightning"
	"github.com/fewsats/blockbuster/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockInvoiceProvider struct {
	mock.Mock
}

func (m *MockInvoiceProvider) CreateInvoice(ctx context.Context,
	priceInUSDCents uint64, currency,
	productName string) (*lightning.LNInvoice, error) {

	args := m.Called(ctx, priceInUSDCents, currency, productName)
	return args.Get(0).(*lightning.LNInvoice), args.Error(1)
}

type MockStore struct {
	mock.Mock
}

func (m *MockStore) CreateRootKey(ctx context.Context,
	tokenID [32]byte, rootKey [32]byte) error {

	args := m.Called(ctx, tokenID, rootKey)
	return args.Error(0)
}

func (m *MockStore) GetRootKey(ctx context.Context,
	tokenID [32]byte) ([32]byte, error) {

	args := m.Called(ctx, tokenID)
	return args.Get(0).([32]byte), args.Error(1)
}

func (m *MockStore) StoreInvoice(ctx context.Context,
	userID uint64, invoice *lightning.LNInvoice) error {

	args := m.Called(ctx, userID, invoice)
	return args.Error(0)
}

// MockRandReader is a mock for the crypto/rand Reader
type MockRandReader struct {
	mock.Mock
}

func (m *MockRandReader) Read(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

// TestNewChallenge tests the NewChallenge method of the Authenticator
func TestNewChallenge(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockLogger := slog.Default()
	mockProvider := new(MockInvoiceProvider)
	mockStore := new(MockStore)
	mockClock := new(utils.MockClock)

	// Mock random generation
	mockRand := new(MockRandReader)
	rand.Reader = mockRand

	// Test cases
	testCases := []struct {
		name            string
		productName     string
		pubKeyHex       string
		priceInUSDCents uint64
		caveats         map[string]string
		setupMocks      func()
		expectedError   string
	}{
		{
			name:            "Happy Path",
			productName:     "Test Product",
			priceInUSDCents: 1000,
			pubKeyHex:       "0101010101010101010101010101010101010101010101010101010101010101",
			caveats:         map[string]string{"key": "value"},
			setupMocks: func() {
				expectedInvoice := &lightning.LNInvoice{
					PaymentHash:    "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					PaymentRequest: "lnbc...",
				}
				mockProvider.On("CreateInvoice", ctx, uint64(1000), "USD",
					"Test Product").Return(expectedInvoice, nil)

				mockRand.On("Read", mock.Anything).Run(
					func(args mock.Arguments) {
						b := args.Get(0).([]byte)
						copy(b, bytes.Repeat([]byte{0x01}, len(b)))
					}).Return(32, nil)

				expectedTokenID := [32]byte{}
				expectedRootKey := [32]byte{}
				for i := range expectedTokenID {
					expectedTokenID[i] = 0x01
					expectedRootKey[i] = 0x01
				}
				mockStore.On("CreateRootKey", ctx, expectedTokenID, expectedRootKey).Return(nil)

			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mocks
			mockProvider.ExpectedCalls = nil
			mockStore.ExpectedCalls = nil
			mockRand.ExpectedCalls = nil

			// Setup mocks
			tc.setupMocks()

			// Create authenticator
			authenticator := NewAuthenticator(mockLogger, mockProvider,
				DefaultConfig(), mockStore, mockClock)

			// Execute
			challenge, err := authenticator.NewChallenge(ctx, tc.productName,
				tc.pubKeyHex, tc.priceInUSDCents, tc.caveats)

			// Assert
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				assert.Nil(t, challenge)
			} else {
				require.NoError(t, err)
				require.NotNil(t, challenge)

				// Verify the invoice
				assert.Equal(t,
					"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					challenge.Invoice.PaymentHash)
				assert.Equal(t, "lnbc...", challenge.Invoice.PaymentRequest)

				// Verify the macaroon
				assert.NotNil(t, challenge.Macaroon)

				// Verify method calls
				mockProvider.AssertExpectations(t)
				mockStore.AssertExpectations(t)
				mockRand.AssertExpectations(t)
			}
		})
	}
}

func generateKeysAndSignature(message string) (pubKeyHex,
	signatureHex string) {
	// Generate a private key
	privKey, err := btcec.NewPrivateKey()
	if err != nil {
		return "", ""
	}

	// Get the corresponding public key
	pubKey := privKey.PubKey()

	// Encode the public key to hex
	pubKeyHex = hex.EncodeToString(schnorr.SerializePubKey(pubKey))

	// Hash the message
	hash := sha256.Sum256([]byte(message))

	// Sign the message hash
	sig, err := schnorr.Sign(privKey, hash[:])
	if err != nil {
		return "", ""
	}

	// Serialize the signature to hex
	signatureHex = hex.EncodeToString(sig.Serialize())

	return pubKeyHex, signatureHex
}

func TestValidateSignature(t *testing.T) {
	domain := "localhost:8080"
	timestamp := time.Now().Unix()
	message := fmt.Sprintf("%s:%d", domain, timestamp)

	pubKeyHex, signatureHex := generateKeysAndSignature(message)
	_, signatureHex2 := generateKeysAndSignature("random message")

	testCases := []struct {
		name          string
		pubKeyHex     string
		signatureHex  string
		domain        string
		timestamp     int64
		expectedError string
	}{
		{
			name:          "valid signature",
			pubKeyHex:     pubKeyHex,
			signatureHex:  signatureHex,
			domain:        domain,
			timestamp:     timestamp,
			expectedError: "",
		},
		{
			name:          "invalid domain",
			pubKeyHex:     pubKeyHex,
			signatureHex:  signatureHex,
			domain:        "invalid.com",
			timestamp:     timestamp,
			expectedError: "invalid domain",
		},
		{
			name:          "old timestamp",
			pubKeyHex:     pubKeyHex,
			signatureHex:  signatureHex,
			domain:        "localhost:8080",
			timestamp:     time.Now().Add(-11 * time.Minute).Unix(),
			expectedError: "timestamp is too old",
		},
		{
			name:          "invalid signature",
			pubKeyHex:     pubKeyHex,
			signatureHex:  signatureHex2,
			domain:        "localhost:8080",
			timestamp:     timestamp,
			expectedError: "signature verification failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authenticator := &Authenticator{cfg: DefaultConfig()}
			err := authenticator.ValidateSignature(tc.pubKeyHex, tc.signatureHex, tc.domain, tc.timestamp)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
