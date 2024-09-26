package video_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/fewsats/blockbuster/l402"
	"github.com/fewsats/blockbuster/lightning"
	"github.com/fewsats/blockbuster/utils"
	"github.com/fewsats/blockbuster/video"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/macaroon.v2"
)

// Mock dependencies
type MockStore struct {
	mock.Mock
}

var mac, _ = macaroon.New([]byte("rootKey"), []byte("id"),
	"location", macaroon.LatestVersion)

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

func (m *MockManager) GenerateStreamURL(ctx context.Context, externalID string) (string, string, error) {
	args := m.Called(ctx, externalID)
	return args.Get(0).(string), args.Get(1).(string), args.Error(2)
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

func (m *MockCloudflareService) GenerateStreamURL(ctx context.Context, externalID string) (string, string, error) {
	args := m.Called(ctx, externalID)
	return args.String(0), args.String(1), args.Error(2)
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

	testCases := []struct {
		name           string
		authHeader     string
		reqBody        *video.StreamVideoRequest
		setupMocks     func(*MockStore, *MockAuthenticator, *MockOrdersMgr, *MockCloudflareService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "video not ready",
			authHeader: "validAuthHeader",
			reqBody:    nil,
			setupMocks: func(
				mockStore *MockStore,
				mockAuthenticator *MockAuthenticator,
				mockOrdersMgr *MockOrdersMgr,
				mockCloudflare *MockCloudflareService,
			) {
				mockStore.On("GetVideoByExternalID", mock.Anything,
					"externalID").Return(&video.Video{ReadyToStream: false}, nil)
				mockCloudflare.On(
					"GetStreamVideoInfo", mock.Anything, "externalID",
				).Return(&cloudflare.StreamVideo{}, errors.New("video not ready"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Video does not exist or is not ready to stream",
		},
		{
			name:       "first video access, video ready in cf, valid credentials",
			authHeader: "validAuthHeader",
			reqBody:    nil,
			setupMocks: func(mockStore *MockStore,
				mockAuthenticator *MockAuthenticator,
				mockOrdersMgr *MockOrdersMgr,
				mockCloudflare *MockCloudflareService,
			) {
				mockStore.On(
					"GetVideoByExternalID", mock.Anything, "externalID",
				).Return(&video.Video{ReadyToStream: false}, nil)
				mockCloudflare.On("GetStreamVideoInfo", mock.Anything,
					"externalID",
				).Return(&cloudflare.StreamVideo{ReadyToStream: true}, nil)
				mockStore.On("UpdateVideo", mock.Anything, "externalID",
					mock.Anything,
				).Return(&video.Video{ReadyToStream: true}, nil)

				mockAuthenticator.On(
					"ValidateL402Credentials", mock.Anything, "validAuthHeader",
				).Return("paymentHash", nil)
				mockOrdersMgr.On(
					"RecordPurchase", mock.Anything, "paymentHash", "videos",
				).Return(nil)
				mockStore.On(
					"IncrementVideoViews", mock.Anything, "externalID",
				).Return(nil)
				mockCloudflare.On(
					"GenerateStreamURL", mock.Anything, "externalID",
				).Return("http://hls_stream.url", "http://dash_stream.url", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "http://hls_stream.url",
		},
		{
			name:       "valid credentials",
			authHeader: "validAuthHeader",
			reqBody:    nil,
			setupMocks: func(mockStore *MockStore,
				mockAuthenticator *MockAuthenticator,
				mockOrdersMgr *MockOrdersMgr,
				mockCloudflare *MockCloudflareService,
			) {
				mockStore.On(
					"GetVideoByExternalID", mock.Anything, "externalID",
				).Return(&video.Video{ReadyToStream: true}, nil)
				mockAuthenticator.On(
					"ValidateL402Credentials", mock.Anything, "validAuthHeader",
				).Return("paymentHash", nil)
				mockOrdersMgr.On(
					"RecordPurchase", mock.Anything, "paymentHash", "videos",
				).Return(nil)
				mockStore.On(
					"IncrementVideoViews", mock.Anything, "externalID",
				).Return(nil)
				mockCloudflare.On(
					"GenerateStreamURL", mock.Anything, "externalID",
				).Return("http://hls_stream.url", "http://dash_stream.url", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "http://hls_stream.url",
		},
		{
			name:       "invalid credentials",
			authHeader: "validAuthHeader",
			reqBody:    nil,
			setupMocks: func(mockStore *MockStore,
				mockAuthenticator *MockAuthenticator,
				mockOrdersMgr *MockOrdersMgr,
				mockCloudflare *MockCloudflareService,
			) {
				mockStore.On(
					"GetVideoByExternalID", mock.Anything, "externalID",
				).Return(&video.Video{ReadyToStream: true}, nil)
				mockAuthenticator.On(
					"ValidateL402Credentials", mock.Anything, "validAuthHeader",
				).Return("", errors.New("unrecoverable formatting error in credentials"))

			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "unable to extract L402 credentials",
		},
		{
			name:       "invalid signature",
			authHeader: "validAuthHeader",
			reqBody: &video.StreamVideoRequest{
				Signature: "invalidSignature",
				Domain:    "invalidDomain",
				Timestamp: 1234567890,
				PubKey:    "invalidPubkey",
			},
			setupMocks: func(mockStore *MockStore,
				mockAuthenticator *MockAuthenticator,
				mockOrdersMgr *MockOrdersMgr,
				mockCloudflare *MockCloudflareService,
			) {
				mockStore.On(
					"GetVideoByExternalID", mock.Anything, "externalID",
				).Return(&video.Video{ReadyToStream: true}, nil)
				mockAuthenticator.On(
					"ValidateL402Credentials", mock.Anything, "validAuthHeader",
				).Return("", l402.ErrMissingAuthorizationHeader)
				mockAuthenticator.On(
					"ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(errors.New("signature is invalid"))

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "signature is invalid",
		},
		{
			name:       "payment required",
			authHeader: "validAuthHeader",
			reqBody: &video.StreamVideoRequest{
				Signature: "validSignature",
				Domain:    "validDomain",
				Timestamp: 1234567890,
				PubKey:    "validPubkey",
			},
			setupMocks: func(mockStore *MockStore,
				mockAuthenticator *MockAuthenticator,
				mockOrdersMgr *MockOrdersMgr,
				mockCloudflare *MockCloudflareService,
			) {
				mockStore.On(
					"GetVideoByExternalID", mock.Anything, "externalID",
				).Return(&video.Video{PriceInCents: 1, Title: "title",
					ExternalID: "externalID", ReadyToStream: true,
					UserID: 661}, nil)
				mockAuthenticator.On(
					"ValidateL402Credentials", mock.Anything, "validAuthHeader",
				).Return("", l402.ErrMissingAuthorizationHeader)
				mockAuthenticator.On(
					"ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil)
				mockAuthenticator.On(
					"NewChallenge", mock.Anything, "title", "validPubkey", uint64(1), mock.Anything,
				).Return(&l402.Challenge{
					Invoice: &lightning.LNInvoice{
						PaymentHash:    "paymentHash",
						PaymentRequest: "paymentRequest",
					},
					Macaroon: mac,
				}, nil)
				mockOrdersMgr.On(
					"CreateOffer", mock.Anything, int64(661), uint64(1), "externalID", "paymentHash",
				).Return(nil)

			},
			expectedStatus: http.StatusPaymentRequired,
			expectedBody:   "Payment Required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStore := new(MockStore)
			mockAuthenticator := new(MockAuthenticator)
			mockOrdersMgr := new(MockOrdersMgr)
			mockCloudflare := new(MockCloudflareService)
			mockLogger := slog.Default()

			manager := video.NewManager(
				mockOrdersMgr,
				mockCloudflare,
				mockAuthenticator,
				mockStore,
				mockLogger,
				utils.NewMockClock(),
			)
			controller := video.NewController(
				manager,
				mockAuthenticator,
				mockStore,
				mockLogger,
				video.DefaultConfig(),
			)

			router := gin.New()
			router.POST("/video/stream/:id", controller.StreamVideo)

			tc.setupMocks(
				mockStore,
				mockAuthenticator,
				mockOrdersMgr,
				mockCloudflare,
			)

			body, err := json.Marshal(tc.reqBody)
			require.NoError(t, err)
			req, err := http.NewRequest(
				http.MethodPost,
				"/video/stream/externalID",
				bytes.NewBuffer(body),
			)
			require.NoError(t, err)
			req.Header.Set("Authorization", tc.authHeader)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedBody)
		})
	}
}
