package video_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/fewsats/blockbuster/l402"
	"github.com/fewsats/blockbuster/utils"
	"github.com/fewsats/blockbuster/video"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock dependencies
type MockStore struct {
	mock.Mock
}

func (m *MockStore) GetOrCreateUserByEmail(ctx context.Context, email string) (int64, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStore) CreateVideo(ctx context.Context, params video.CreateVideoParams) (*video.Video, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*video.Video), args.Error(1)
}

func (m *MockStore) ListUserVideos(ctx context.Context, userID int64) ([]*video.Video, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*video.Video), args.Error(1)
}

func (m *MockStore) GetVideoByExternalID(ctx context.Context, externalID string) (*video.Video, error) {
	args := m.Called(ctx, externalID)
	return args.Get(0).(*video.Video), args.Error(1)
}

func (m *MockStore) IncrementVideoViews(ctx context.Context, externalID string) error {
	args := m.Called(ctx, externalID)
	return args.Error(0)
}

func (m *MockStore) UpdateVideo(ctx context.Context, externalID string, params *video.CloudflareVideoInfo) (*video.Video, error) {
	args := m.Called(ctx, externalID, params)
	return args.Get(0).(*video.Video), args.Error(1)
}

type MockAuthenticator struct {
	mock.Mock
}

func (m *MockAuthenticator) ValidateSignature(pubKey, signature, domain string, timestamp int64) error {
	args := m.Called(pubKey, signature, domain, timestamp)
	return args.Error(0)
}

func (m *MockAuthenticator) ValidateL402Credentials(ctx context.Context, authHeader string) (string, error) {
	args := m.Called(ctx, authHeader)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockAuthenticator) NewChallenge(ctx context.Context, domain, pubKeyHex string,
	priceInCents uint64, caveats map[string]string) (*l402.Challenge, error) {
	args := m.Called(ctx, domain, pubKeyHex, priceInCents, caveats)
	return args.Get(0).(*l402.Challenge), args.Error(1)
}

type MockManager struct {
	mock.Mock
}

func (m *MockManager) IsVideoReady(ctx context.Context, externalID string) error {
	args := m.Called(ctx, externalID)
	return args.Error(0)
}

func (m *MockManager) RecordPurchase(ctx context.Context, externalID,
	paymentHash, resource string) error {

	args := m.Called(ctx, externalID, paymentHash, resource)
	return args.Error(0)
}

func (m *MockManager) GenerateStreamURL(ctx context.Context, externalID string) (string, error) {
	args := m.Called(ctx, externalID)
	return args.Get(0).(string), args.Error(1)
}

type MockOrdersMgr struct {
	mock.Mock
}

func (m *MockOrdersMgr) CreateOffer(ctx context.Context, userID int64, priceInCents uint64, externalID, paymentHash string) error {
	args := m.Called(ctx, userID, priceInCents, externalID, paymentHash)
	return args.Error(0)
}

func (m *MockOrdersMgr) RecordPurchase(ctx context.Context, payHash, serviceType string) error {
	args := m.Called(ctx, payHash, serviceType)
	return args.Error(0)
}

type MockCloudflareService struct {
	mock.Mock
}

func (m *MockCloudflareService) GenerateStreamURL(ctx context.Context, externalID string) (string, error) {
	args := m.Called(ctx, externalID)
	return args.String(0), args.Error(1)
}

func (m *MockCloudflareService) GenerateVideoUploadURL(ctx context.Context) (string, string, error) {
	args := m.Called(ctx)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockCloudflareService) UploadPublicFile(ctx context.Context, key, prefix string,
	reader io.ReadSeeker) (string, error) {

	args := m.Called(ctx, key, prefix, reader)
	return args.String(0), args.Error(1)
}

func (m *MockCloudflareService) GetStreamVideoInfo(ctx context.Context, externalID string) (*cloudflare.StreamVideo, error) {
	args := m.Called(ctx, externalID)
	return args.Get(0).(*cloudflare.StreamVideo), args.Error(1)
}

func TestStreamVideo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockStore := new(MockStore)
	mockAuthenticator := new(MockAuthenticator)
	mockOrdersMgr := new(MockOrdersMgr)
	mockCloudflare := new(MockCloudflareService)
	mockLogger := slog.Default()

	controller := video.NewController(mockCloudflare, mockOrdersMgr, mockAuthenticator, mockStore, mockLogger, utils.NewMockClock())

	router := gin.New()
	router.POST("/video/stream/:id", controller.StreamVideo)

	// Define request bodies using structs
	// reqBody := video.StreamVideoRequest{
	// 	PubKey:    "pubKey",
	// 	Domain:    "domain",
	// 	Timestamp: time.Now().Unix(),
	// 	Signature: "signature",
	// }

	testCases := []struct {
		name           string
		authHeader     string
		reqBody        *video.StreamVideoRequest
		setupMocks     func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "valid credentials",
			authHeader: "validAuthHeader",
			reqBody:    nil,
			setupMocks: func() {
				mockAuthenticator.On("ValidateSignature", "pubKey", "signature", "domain", mock.Anything).Return(nil)
				mockAuthenticator.On("ValidateL402Credentials", mock.Anything, "validAuthHeader").Return("paymentHash", nil)
				mockOrdersMgr.On("RecordPurchase", mock.Anything, "paymentHash", "videos").Return(nil)
				mockOrdersMgr.On("IncrementVideoViews", mock.Anything, "externalID").Return(nil)
				mockStore.On("GetVideoByExternalID", mock.Anything, "externalID").Return(&video.Video{ReadyToStream: true}, nil)
				mockStore.On("IncrementVideoViews", mock.Anything, "externalID").Return(nil)
				mockCloudflare.On("GenerateStreamURL", mock.Anything, "externalID").Return("http://stream.url", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "http://stream.url",
		},
		// {
		// 	name:       "invalid credentials, payment required",
		// 	authHeader: "invalidAuthHeader",
		// 	setupMocks: func() {
		// 		mockAuthenticator.On("ValidateSignature", "pubKey", "signature", "domain", mock.Anything).Return(nil)
		// 		mockAuthenticator.On("ValidateL402Credentials", mock.Anything, "invalidAuthHeader").Return("", l402.ErrInvalidPreimage)
		// 		mockStore.On("GetVideoByExternalID", mock.Anything, "externalID").Return(&video.Video{ReadyToStream: true}, nil)
		// 	},
		// 	expectedStatus: http.StatusPaymentRequired,
		// 	expectedBody:   "Payment Required",
		// },
		// {
		// 	name:       "invalid request",
		// 	authHeader: "invalidAuthHeader",
		// 	setupMocks: func() {
		// 		mockAuthenticator.On("ValidateSignature", "pubKey", "signature", "domain", mock.Anything).Return(nil)
		// 		mockAuthenticator.On("ValidateL402Credentials", mock.Anything, "invalidAuthHeader").Return("", errors.New("invalid request"))
		// 		mockManager.On("IsVideoReady", mock.Anything, "externalID").Return(nil)
		// 		mockStore.On("GetVideoByExternalID", mock.Anything, "externalID").Return(&video.Video{ReadyToStream: true}, nil)
		// 	},
		// 	expectedStatus: http.StatusBadRequest,
		// 	expectedBody:   "invalid request",
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			body, err := json.Marshal(tc.reqBody)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/video/stream/externalID", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Authorization", tc.authHeader)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedBody)
		})
	}
}
