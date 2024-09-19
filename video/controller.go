package video

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"

	"log/slog"

	"github.com/fewsats/blockbuster/cloudflare"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	cf *cloudflare.Service

	store  Store
	logger *slog.Logger
}

func NewController(cloudflareService *cloudflare.Service, store Store,
	logger *slog.Logger) *Controller {

	return &Controller{
		cf: cloudflareService,

		store:  store,
		logger: logger,
	}
}

func (c *Controller) RegisterPublicRoutes(router *gin.Engine) {
	router.POST("/video/upload", c.UploadVideo)
	router.GET("/video/:id", nil)
	router.GET("/video/search", nil)
}

func (c *Controller) RegisterProtectedRoutes(router *gin.Engine) {
	router.GET("/user/videos", nil)
	router.DELETE("/video/:id", nil)
}

type UploadVideoRequest struct {
	Email        string `form:"email" binding:"required,email"`
	Title        string `form:"title" binding:"required"`
	Description  string `form:"description"`
	PriceInCents int64  `form:"price_in_cents" binding:"required,min=0"`
	CoverImage   string `form:"cover_image"`
}

// UploadFileResponse represents a response to uploading a file.
type UploadVideoResponse struct {
	VideoID   string `json:"video_id"`
	UploadURL string `json:"upload_url"`
}

func (r *UploadVideoRequest) Validate() error {
	if r.CoverImage == "" {
		return fmt.Errorf("cover image is required")
	}
	return nil
}
func (c *Controller) processAndUploadCoverImage(externalID string,
	coverImageData string) (string, error) {

	coverImageBytes, err := base64.StdEncoding.DecodeString(coverImageData)
	if err != nil {
		return "", fmt.Errorf("failed to decode cover image data: %w", err)
	}

	coverImageReader := bytes.NewReader(coverImageBytes)

	coverURL, err := c.cf.UploadPublicFile(externalID, "cover-images", coverImageReader)
	if err != nil {
		return "", fmt.Errorf("failed to upload cover file: %w", err)
	}

	return coverURL, nil
}

func (c *Controller) UploadVideo(ctx *gin.Context) {
	var req UploadVideoRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	videoID := uuid.New().String()

	coverURL, err := c.processAndUploadCoverImage(
		videoID, req.CoverImage,
	)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Failed to upload cover image"},
		)
		return
	}

	uploadURL, err := c.cf.GenerateVideoUploadURL(videoID)
	if err != nil {
		c.logger.Error("Failed to generate upload URL", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate upload URL"})
		return
	}

	_, err = c.store.CreateVideo(ctx, CreateVideoParams{
		ExternalID:   videoID,
		UserEmail:    req.Email,
		Title:        req.Title,
		Description:  req.Description,
		CoverURL:     coverURL,
		VideoURL:     uploadURL,
		PriceInCents: req.PriceInCents,
	})

	if err != nil {
		c.logger.Error("Failed to create video", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video metadata"})
		return
	}

	ctx.JSON(
		http.StatusOK,
		UploadVideoResponse{
			VideoID:   videoID,
			UploadURL: uploadURL,
		},
	)
}
