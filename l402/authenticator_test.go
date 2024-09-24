package l402

import (
	"bytes"
	"context"
	"crypto/rand"
	"log/slog"
	"testing"
	"time"

	"github.com/fewsats/blockbuster/lightning"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/macaroon.v2"
)

type MockInvoiceProvider struct {
	mock.Mock
}

func (m *MockInvoiceProvider) CreateInvoice(ctx context.Context, priceInUSDCents uint64, currency, productName string) (*lightning.LNInvoice, error) {
	args := m.Called(ctx, priceInUSDCents, currency, productName)
	return args.Get(0).(*lightning.LNInvoice), args.Error(1)
}

type MockMacaroonManager struct {
	mock.Mock
}

func (m *MockMacaroonManager) MintMacaroon(ctx context.Context, location string, identifier, rootKey []byte, caveats map[string]string) (*macaroon.Macaroon, error) {
	args := m.Called(ctx, location, identifier, rootKey, caveats)
	return args.Get(0).(*macaroon.Macaroon), args.Error(1)
}

func (m *MockMacaroonManager) ValidateMacaroon(ctx context.Context,
	authContext map[string]string, macaroonStr string) (ValidationResult, error) {

	args := m.Called(ctx, authContext, macaroonStr)
	return args.Get(0).(ValidationResult), args.Error(1)
}

func (m *MockMacaroonManager) EncodeMacaroon(mac *macaroon.Macaroon) (string, error) {
	args := m.Called(mac)
	return args.String(0), args.Error(1)
}

type MockStore struct {
	mock.Mock
}

func (m *MockStore) CreateRootKey(ctx context.Context, tokenID [32]byte, rootKey [32]byte) error {
	args := m.Called(ctx, tokenID, rootKey)
	return args.Error(0)
}

func (m *MockStore) GetRootKey(ctx context.Context, tokenID [32]byte) ([32]byte, error) {
	args := m.Called(ctx, tokenID)
	return args.Get(0).([32]byte), args.Error(1)
}

func (m *MockStore) StoreInvoice(ctx context.Context, userID uint64, invoice *lightning.LNInvoice) error {
	args := m.Called(ctx, userID, invoice)
	return args.Error(0)
}

type MockClock struct {
	mock.Mock
}

func (m *MockClock) Now() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
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
	mockClock := new(MockClock)
	mockMacaroons := new(MockMacaroonManager)

	// Mock random generation
	mockRand := new(MockRandReader)
	rand.Reader = mockRand

	// Test cases
	testCases := []struct {
		name            string
		productName     string
		priceInUSDCents uint64
		caveats         map[string]string
		setupMocks      func()
		expectedError   string
	}{
		{
			name:            "Happy Path",
			productName:     "Test Product",
			priceInUSDCents: 1000,
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

				mockMacaroons.On("MintMacaroon", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&macaroon.Macaroon{}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mocks
			mockProvider.ExpectedCalls = nil
			mockStore.ExpectedCalls = nil
			mockRand.ExpectedCalls = nil
			mockMacaroons.ExpectedCalls = nil

			// Setup mocks
			tc.setupMocks()

			// Create authenticator
			authenticator := NewAuthenticator(mockLogger, mockProvider, mockMacaroons, mockStore, mockClock)

			// Execute
			challenge, err := authenticator.NewChallenge(ctx, tc.productName, tc.priceInUSDCents, tc.caveats)

			// Assert
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				assert.Nil(t, challenge)
			} else {
				require.NoError(t, err)
				require.NotNil(t, challenge)

				// Verify the invoice
				assert.Equal(t, "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", challenge.Invoice.PaymentHash)
				assert.Equal(t, "lnbc...", challenge.Invoice.PaymentRequest)

				// Verify the macaroon
				assert.NotNil(t, challenge.Macaroon)

				// Verify method calls
				mockProvider.AssertExpectations(t)
				mockStore.AssertExpectations(t)
				mockRand.AssertExpectations(t)
				mockMacaroons.AssertExpectations(t)
			}
		})
	}
}
