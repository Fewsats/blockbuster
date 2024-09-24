package video

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"log/slog"

	"github.com/fewsats/blockbuster/cloudflare"
	"github.com/fewsats/blockbuster/l402"
	"github.com/fewsats/blockbuster/utils"
	"github.com/gin-gonic/gin"
)

const (
	// ExpirationTime is the time after which a macaroon expires.
	ExpirationTime = 24 * time.Hour * 30 * 12 // 12 months
)

type Controller struct {
	cf            *cloudflare.Service
	videos        *Manager
	authenticator Authenticator

	store  Store
	logger *slog.Logger
}

func NewController(cf CloudflareService, orders OrdersMgr,
	authenticator Authenticator, store Store, logger *slog.Logger,
	clock utils.Clock) *Controller {

	manager := NewManager(orders, cf, authenticator,
		store, logger, clock)

	return &Controller{
		authenticator: authenticator,
		videos:        manager,

		store:  store,
		logger: logger,
	}
}

func (c *Controller) RegisterPublicRoutes(router *gin.Engine) {
	router.POST("/video/upload", c.UploadVideo)
	// router.GET("/video/:id", c.ServeVideoPage)
	router.GET("/video/search", nil)

}

func (c *Controller) RegisterL402Routes(router *gin.Engine) {
	router.POST("/video/stream/:id", c.StreamVideo)
}

func (c *Controller) RegisterProtectedRoutes(router *gin.Engine) {
	router.GET("/user/videos", c.ListUserVideos)
}

type UploadVideoRequest struct {
	Email        string                `form:"email" binding:"required,email"`
	Title        string                `form:"title" binding:"required"`
	Description  string                `form:"description"`
	PriceInCents int64                 `form:"price_in_cents" binding:"required,min=0"`
	CoverImage   *multipart.FileHeader `form:"cover_image"`
}

// UploadFileResponse represents a response to uploading a file.
type UploadVideoResponse struct {
	VideoID   string `json:"video_id"`
	UploadURL string `json:"upload_url"`
}

func (r *UploadVideoRequest) Validate() error {
	if r.CoverImage == nil {
		return fmt.Errorf("cover image is required")
	}
	return nil
}

func (c *Controller) processAndUploadCoverImage(gCtx context.Context,
	externalID string, coverImageHeader *multipart.FileHeader) (string, error) {

	file, err := coverImageHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open cover image: %w", err)
	}
	defer file.Close()

	// Read the entire file into memory
	coverImageBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read cover image: %w", err)
	}

	coverImageReader := bytes.NewReader(coverImageBytes)

	coverURL, err := c.cf.UploadPublicFile(gCtx, externalID,
		"cover-images", coverImageReader)
	if err != nil {
		return "", fmt.Errorf("failed to upload cover file: %w", err)
	}

	return coverURL, nil
}

func (c *Controller) UploadVideo(gCtx *gin.Context) {
	var req UploadVideoRequest
	if err := gCtx.ShouldBind(&req); err != nil {
		c.logger.Error("Invalid request", "error", err)
		gCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.logger.Error("Invalid request", "error", err)
		gCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := c.store.GetOrCreateUserByEmail(gCtx, req.Email)
	if err != nil {
		c.logger.Error("Failed to get user by email", "error", err)
		gCtx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Failed to get user"},
		)
		return
	}

	uploadURL, externalID, err := c.cf.GenerateVideoUploadURL(gCtx)
	if err != nil {
		c.logger.Error("Failed to generate upload URL", "error", err)
		gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate upload URL"})
		return
	}

	coverURL, err := c.processAndUploadCoverImage(gCtx, externalID, req.CoverImage)
	if err != nil {
		c.logger.Error("Failed to upload cover image", "error", err)
		gCtx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Failed to upload cover image"},
		)
		return
	}
	_, err = c.store.CreateVideo(gCtx, CreateVideoParams{
		ExternalID:   externalID,
		UserID:       userID,
		Title:        req.Title,
		Description:  req.Description,
		CoverURL:     coverURL,
		PriceInCents: req.PriceInCents,
	})

	if err != nil {
		c.logger.Error("Failed to create video", "error", err)
		gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video metadata"})
		return
	}

	gCtx.JSON(
		http.StatusOK,
		UploadVideoResponse{
			VideoID:   externalID,
			UploadURL: uploadURL,
		},
	)
}

func (c *Controller) ListUserVideos(gCtx *gin.Context) {
	userID := gCtx.GetInt64("user_id")
	if userID == 0 {
		gCtx.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "User not authenticated"},
		)
		return
	}

	videos, err := c.store.ListUserVideos(gCtx, userID)
	if err != nil {
		c.logger.Error("Failed to list user videos", "error", err)
		gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user videos"})
		return
	}

	// TODO(pol): this needs to be the L402 URL link
	// for _, v := range videos {
	// 	v.CoverURL = c.cf.PublicFileURL(fmt.Sprintf("cover-images/%s", v.ExternalID))
	// }

	gCtx.JSON(http.StatusOK, gin.H{"videos": videos})
}

// extractExternalVideoID extracts the external video ID from the request.
func extractExternalVideoID(gCtx *gin.Context) (string, error) {
	externalID := gCtx.Param("id")

	externalID = strings.TrimSuffix(externalID, "/")
	externalID = strings.TrimPrefix(externalID, "/")

	return externalID, nil
}

type StreamVideoRequest struct {
	PubKey    string `json:"pub_key" binding:"required"`
	Domain    string `json:"domain" binding:"required"`
	Timestamp int64  `json:"timestamp" binding:"required"`
	Signature string `json:"signature" binding:"required"`
}

// func (r *StreamVideoRequest) Validate(auth *l402.Authenticator) error {
// 	if time.Since(time.Unix(r.Timestamp, 0)) > 10*time.Minute ||
// 		time.Until(time.Unix(r.Timestamp, 0)) > 10*time.Minute {

// 		return fmt.Errorf("timestamp is too old or too far in the future")
// 	}

// 	// TODO(pol): validate domain
// 	if r.Domain != "ourdomain.com" {
// 		return fmt.Errorf("invalid domain")
// 	}

// 	return nil
// }

func (c *Controller) StreamVideo(gCtx *gin.Context) {
	ctx := gCtx.Request.Context()
	externalID, err := extractExternalVideoID(gCtx)
	if err != nil {
		c.logger.Debug(
			"invalid file ID",
			"error", err,
		)
		gCtx.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid file ID"},
		)

		return
	}

	// Step 0: Check if the video is ready to stream
	err = c.videos.IsVideoReady(ctx, externalID)
	if err != nil {
		c.logger.Error("Video is not ready to stream", "error", err)
		gCtx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Video does not exist or is not ready to stream"},
		)
		return
	}

	// Step 1: Check if the user has provided valid L402 credentials
	autHeader := gCtx.GetHeader("Authorization")
	paymentHash, err := c.authenticator.ValidateL402Credentials(ctx, autHeader)

	switch {
	// Step 1.1: A set of valid L402 credentials was provided so we will record the sale
	// and allow the user to stream the video.
	case err == nil:
		err = c.videos.RecordPurchaseAndView(ctx, externalID, paymentHash, "videos")
		if err != nil {
			c.logger.Error(
				"failed to record purchase",
				"external_id", externalID,
				"error", err,
			)
		}

		presignedURL, err := c.videos.GenerateStreamURL(gCtx, externalID)
		if err != nil {
			c.logger.Error("Failed to generate stream URL", "error", err)
			gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate stream URL"})
			return
		}

		gCtx.JSON(http.StatusOK, gin.H{"url": presignedURL})
		return

	case !errors.Is(err, l402.ErrMissingAuthorizationHeader) && !errors.Is(err, l402.ErrInvalidPreimage):
		// Step 1.2 Unexpected error with L402 credentials, we will fail instead of returning a L402
		c.logger.Debug(
			"unable to extract L402 credentials",
			"header", autHeader,
			"error", err,
		)
		gCtx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "unable to extract L402 credentials"},
		)

		return
	}

	// Step 2: No L402 credentials have been provided, let's check if the request is from a valid
	// user and contains all the required params to send back a challenge
	// This step is reached when err == ErrMissingAuthorizationHeader || ErrInvalidPreimage
	c.logger.Debug(
		"missing Authorization header",
		"error", err,
	)

	var req StreamVideoRequest
	if err := gCtx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Invalid request", "error", err)
		gCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = c.authenticator.ValidateSignature(req.PubKey, req.Signature,
		req.Domain, req.Timestamp)
	if err != nil {
		c.logger.Error("Invalid request", "error", err)
		gCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	challenge, err := c.videos.CreateL402Challenge(ctx, req.PubKey, externalID)
	if err != nil {
		c.logger.Error(
			"failed to create challenge",
			"file_id", externalID,
			"error", err,
		)

		gCtx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to create L402 challenge"},
		)

		return
	}

	headerKey := challenge.HeaderKey()
	headerValue, err := challenge.HeaderValue()
	if err != nil {
		c.logger.Error(
			"failed to get challenge header value",
			"error", err,
		)

		gCtx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to get challenge header value"},
		)

		return
	}

	gCtx.Writer.Header().Set(headerKey, headerValue)

	gCtx.JSON(
		http.StatusPaymentRequired,
		gin.H{"error": "Payment Required"},
	)
}
