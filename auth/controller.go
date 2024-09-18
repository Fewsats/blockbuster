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
	ID       int64
	Email    string
	Verified bool
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
	err = c.store.StoreToken(req.Email, token, expiration)
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

func (c *Controller) VerifyTokenHandler(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	email, err := c.store.VerifyToken(token)
	if err != nil {
		c.logger.Error("Failed to verify token", "error", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Generate a new session
	session := sessions.Default(ctx)
	session.Set("user", email)
	if err := session.Save(); err != nil {
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
		user := session.Get("user")
		if user == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			ctx.Abort()
			return
		}
		ctx.Set("email", user)
		ctx.Next()
	}
}

func (c *Controller) MeHandler(ctx *gin.Context) {
	session := sessions.Default(ctx)
	userEmail := session.Get("user")
	if userEmail == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"email": userEmail})
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
	user := session.Get("user")
	if user == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session token"})
		return
	}
	session.Delete("user")
	if err := session.Save(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
