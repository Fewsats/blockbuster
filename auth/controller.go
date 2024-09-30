package auth

import (
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"
	"time"

	"github.com/fewsats/blockbuster/email"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	emailService *email.ResendService
	cfg          *Config

	store  Store
	logger *slog.Logger
}

type User struct {
	ID               int64  `json:"id"`
	Email            string `json:"email"`
	LightningAddress string `json:"lightning_address"`
	Verified         bool   `json:"verified"`
}

type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func NewController(emailService *email.ResendService, logger *slog.Logger,
	store Store, cfg *Config) *Controller {

	return &Controller{
		emailService: emailService,
		cfg:          cfg,

		logger: logger,
		store:  store,
	}
}

// RegisterPublicRoutes registers the public authentication routes.
func (c *Controller) RegisterPublicRoutes(router *gin.Engine) {
	router.POST("/auth/login", c.LoginHandler)
	router.GET("/auth/verify", c.VerifyTokenHandler)
	router.GET("/auth/logout", c.LogoutHandler)
}

// RegisterProtectedRoutes registers the protected authentication routes.
func (c *Controller) RegisterProtectedRoutes(router *gin.Engine) {
	router.GET("/me", c.MeHandler)
	router.POST("/auth/profile", c.UpdateProfileHandler)
}

// RegisterAuthMiddleware registers the authentication middleware.
func (c *Controller) RegisterAuthMiddleware(router *gin.Engine) {
	router.Use(c.AuthMiddleware())
}

func (c *Controller) LoginHandler(ctx *gin.Context) {
	var req LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	token, err := generateRandomToken(32)
	if err != nil {
		c.logger.Error("Failed to generate random token", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	// TODO(maybe set higher expiration so we can include here a "welcome" email?)
	expiration := time.Now().Add(time.Duration(c.cfg.TokenExpirationMinutes) * time.Minute)
	err = c.store.StoreToken(ctx, req.Email, token, expiration)
	if err != nil {
		c.logger.Error("Failed to store token", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process login request"})
		return
	}

	err = c.emailService.SendMagicLink(req.Email, token)
	if err != nil {
		c.logger.Error("Failed to send email", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send magic link email. Please try again later."})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Magic link sent to your email. Please check your inbox."})
}

// VerifyTokenHandler verifies the user token sent with a magic link
// and sets the user session.
func (c *Controller) VerifyTokenHandler(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	email, err := c.store.VerifyToken(ctx, token)
	if err != nil {
		c.logger.Error("Failed to verify token", "error", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	userID, err := c.store.GetOrCreateUserByEmail(ctx, email)
	if err != nil {
		c.logger.Error("Failed to get user by email", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	err = c.store.UpdateUserVerified(ctx, email, true)
	if err != nil {
		c.logger.Error("Failed to update user verified", "error", err)
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Failed to update user verified"},
		)
		return
	}

	// Generate a new session
	session := sessions.Default(ctx)
	session.Set("user_id", userID)
	session.Options(sessions.Options{
		Path:   "/",
		MaxAge: 365 * 24 * 60 * 60, // 1 year in secs
	})
	err = session.Save()
	if err != nil {
		c.logger.Error("Failed to save session", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Redirect to the homepage
	ctx.Redirect(http.StatusFound, "/")
}

func (c *Controller) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userID := session.Get("user_id")
		if userID == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			ctx.Abort()
			return
		}
		ctx.Set("user_id", userID)
		ctx.Next()
	}
}

func (c *Controller) MeHandler(ctx *gin.Context) {
	session := sessions.Default(ctx)
	userID := session.Get("user_id")
	if userID == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userIDInt, ok := userID.(int64)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := c.store.GetUserByID(ctx, userIDInt)
	if err != nil {
		c.logger.Error("Failed to get user by ID", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// TODO(pol) return also the lightning address
	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (c *Controller) LogoutHandler(ctx *gin.Context) {
	session := sessions.Default(ctx)
	user := session.Get("user_id")
	if user == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session token"})
		return
	}
	session.Delete("user_id")

	if err := session.Save(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func (c *Controller) UpdateProfileHandler(ctx *gin.Context) {
	session := sessions.Default(ctx)
	userID := session.Get("user_id")
	if userID == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userIDInt, ok := userID.(int64)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		LightningAddress string `json:"lightningAddress"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := c.store.UpdateUserLightningAddress(ctx, userIDInt, req.LightningAddress)
	if err != nil {
		c.logger.Error("Failed to update user lightning address", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}
