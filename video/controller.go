package video

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fewsats/blockbuster/database"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	db            *database.SQLiteDB
	uploadPath    string
	thumbnailPath string
}

func NewController(db *database.SQLiteDB, uploadPath, thumbnailPath string) *Controller {
	return &Controller{
		db:            db,
		uploadPath:    uploadPath,
		thumbnailPath: thumbnailPath,
	}
}

func (c *Controller) UploadHandler(ctx *gin.Context) {
	userEmail, _ := ctx.Get("email")
	title := ctx.PostForm("title")
	description := ctx.PostForm("description")

	// Handle video file upload
	videoFile, err := ctx.FormFile("video")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No video file provided"})
		return
	}

	videoFileName := filepath.Join(c.uploadPath, fmt.Sprintf("%d_%s", time.Now().UnixNano(), videoFile.Filename))
	if err := ctx.SaveUploadedFile(videoFile, videoFileName); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video file"})
		return
	}

	// Handle thumbnail upload (optional)
	var thumbnailFileName string
	thumbnailFile, err := ctx.FormFile("thumbnail")
	if err == nil {
		thumbnailFileName = filepath.Join(c.thumbnailPath, fmt.Sprintf("%d_%s", time.Now().UnixNano(), thumbnailFile.Filename))
		if err := ctx.SaveUploadedFile(thumbnailFile, thumbnailFileName); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save thumbnail file"})
			return
		}
	}

	// Update to include price
	priceStr := ctx.PostForm("price")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price"})
		return
	}

	// Save video metadata to database
	videoID, err := c.db.CreateVideo(userEmail.(string), title, description, videoFileName, thumbnailFileName, price)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video metadata"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Video uploaded successfully", "videoID": videoID})
}

func (c *Controller) GetVideoHandler(ctx *gin.Context) {
	videoID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	video, err := c.db.GetVideo(videoID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":            video.ID,
		"userEmail":     video.UserEmail,
		"title":         video.Title,
		"description":   video.Description,
		"thumbnailPath": video.ThumbnailPath,
		"createdAt":     video.CreatedAt,
	})
}

func (c *Controller) ServeVideoHandler(ctx *gin.Context) {
	videoID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	video, err := c.db.GetVideo(videoID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	ctx.File(video.FilePath)
}

type UserVideo struct {
	ID         int64   `json:"id"`
	Title      string  `json:"title"`
	Price      float64 `json:"price"`
	TotalViews int64   `json:"totalViews"`
}

func (c *Controller) GetUserVideos(ctx *gin.Context) {
	userEmail, _ := ctx.Get("email")
	userEmailStr, ok := userEmail.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user email"})
		return
	}

	videos, err := c.db.GetUserVideos(userEmailStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user videos"})
		return
	}

	userVideos := make([]UserVideo, len(videos))
	for i, v := range videos {
		userVideos[i] = UserVideo{
			ID:         v.ID,
			Title:      v.Title,
			Price:      v.Price,
			TotalViews: v.TotalViews,
		}
	}

	ctx.JSON(http.StatusOK, userVideos)
}
