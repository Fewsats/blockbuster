package video

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"

	"log/slog"

	"github.com/fewsats/blockbuster/cloudflare"
	"github.com/gin-gonic/gin"
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
	router.GET("/video/:id", c.ServeVideoPage)
	router.GET("/video/search", nil)
}

func (c *Controller) RegisterProtectedRoutes(router *gin.Engine) {
	router.GET("/user/videos", c.ListUserVideos)
	router.DELETE("/video/:id", c.DeleteVideo)
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

func (c *Controller) processAndUploadCoverImage(externalID string,
	coverImageHeader *multipart.FileHeader) (string, error) {

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

	coverURL, err := c.cf.UploadPublicFile(externalID, "cover-images", coverImageReader)
	if err != nil {
		return "", fmt.Errorf("failed to upload cover file: %w", err)
	}

	return coverURL, nil
}

func (c *Controller) UploadVideo(ctx *gin.Context) {
	var req UploadVideoRequest
	if err := ctx.ShouldBind(&req); err != nil {
		c.logger.Error("Invalid request", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.logger.Error("Invalid request", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := c.store.GetOrCreateUserByEmail(ctx, req.Email)
	if err != nil {
		c.logger.Error("Failed to get user by email", "error", err)
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Failed to get user"},
		)
		return
	}

	uploadURL, streamID, err := c.cf.GenerateVideoUploadURL(ctx)
	if err != nil {
		c.logger.Error("Failed to generate upload URL", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate upload URL"})
		return
	}

	coverURL, err := c.processAndUploadCoverImage(
		streamID, req.CoverImage,
	)
	if err != nil {
		c.logger.Error("Failed to upload cover image", "error", err)
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Failed to upload cover image"},
		)
		return
	}
	_, err = c.store.CreateVideo(ctx, CreateVideoParams{
		ExternalID:   streamID,
		UserID:       userID,
		Title:        req.Title,
		Description:  req.Description,
		CoverURL:     coverURL,
		VideoURL:     c.cf.VideoURL(streamID), // TODO(pol) check if correct after using streams
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
			VideoID:   streamID,
			UploadURL: uploadURL,
		},
	)
}

func (c *Controller) ListUserVideos(ctx *gin.Context) {
	userID := ctx.GetInt64("user_id")
	if userID == 0 {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "User not authenticated"},
		)
		return
	}

	videos, err := c.store.ListUserVideos(ctx, userID)
	if err != nil {
		c.logger.Error("Failed to list user videos", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user videos"})
		return
	}

	for _, v := range videos {
		v.CoverURL = c.cf.PublicFileURL(fmt.Sprintf("cover-images/%s", v.ExternalID))
	}

	ctx.JSON(http.StatusOK, gin.H{"videos": videos})
}

func (c *Controller) DeleteVideo(ctx *gin.Context) {
	videoID := ctx.Param("id")

	err := c.store.DeleteVideo(ctx, videoID)
	if err != nil {
		c.logger.Error("Failed to delete video", "error", err)
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Failed to delete video"},
		)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *Controller) ServeVideoPage(ctx *gin.Context) {
	videoID := ctx.Param("id")

	video, err := c.store.GetVideoByExternalID(ctx, videoID)
	if err != nil {
		c.logger.Error("Failed to get video by ID", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch video"})
		return
	}
	presignedURL, err := c.cf.GenerateVideoViewURL(video.ExternalID)
	if err != nil {
		c.logger.Error("Failed to generate presigned URL", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate video URL"})
		return
	}
	video.VideoURL = presignedURL

	c.logger.Info("Video data", "video", video)

	tmpl, err := template.ParseFiles("frontend/video.html")
	if err != nil {
		c.logger.Error("Failed to parse template", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse template"})
		return
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, video); err != nil {
		c.logger.Error("Failed to execute template", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render template"})
		return
	}

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}
