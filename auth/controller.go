package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/fewsats/blockbuster/email"
	"github.com/fewsats/blockbuster/l402"
	"github.com/fewsats/blockbuster/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	emailService    *email.ResendService
	invoiceProvider l402.InvoiceProvider

	clock  utils.Clock
	cfg    *Config
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

const (
	invoiceCheckInterval = 30 * time.Second
)

func NewController(emailService *email.ResendService,
	invoiceProvider l402.InvoiceProvider, logger *slog.Logger,
	store Store, clock utils.Clock, cfg *Config) *Controller {

	return &Controller{
		emailService:    emailService,
		invoiceProvider: invoiceProvider,

		clock:  clock,
		cfg:    cfg,
		logger: logger,
		store:  store,
	}
}

// RegisterPublicRoutes registers the public authentication routes.
func (c *Controller) RegisterPublicRoutes(router *gin.Engine) {
	router.POST("/auth/login", c.LoginHandler)
	router.GET("/auth/verify", c.VerifyTokenHandler)
	router.GET("/auth/logout", c.LogoutHandler)
	router.GET("/auth/check-invoice/:payment_hash", c.GetInvoicePreimage)
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

func (c *Controller) LoginHandler(gCtx *gin.Context) {
	var req LoginRequest

	if err := gCtx.ShouldBindJSON(&req); err != nil {
		gCtx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	token, err := generateRandomToken(32)
	if err != nil {
		c.logger.Error("Failed to generate random token", "error", err)
		gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	// TODO(maybe set higher expiration so we can include here a "welcome" email?)
	expiration := time.Now().Add(time.Duration(c.cfg.TokenExpirationMinutes) * time.Minute)
	err = c.store.StoreToken(gCtx, req.Email, token, expiration)
	if err != nil {
		c.logger.Error("Failed to store token", "error", err)
		gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process login request"})
		return
	}

	err = c.emailService.SendMagicLink(req.Email, token)
	if err != nil {
		c.logger.Error("Failed to send email", "error", err)
		gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send magic link email. Please try again later."})
		return
	}

	gCtx.JSON(http.StatusOK, gin.H{"message": "Magic link sent to your email. Please check your inbox."})
}

// VerifyTokenHandler verifies the user token sent with a magic link
// and sets the user session.
func (c *Controller) VerifyTokenHandler(gCtx *gin.Context) {
	token := gCtx.Query("token")
	if token == "" {
		gCtx.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	email, err := c.store.VerifyToken(gCtx, token)
	if err != nil {
		c.logger.Error("Failed to verify token", "error", err)
		gCtx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	userID, err := c.store.GetOrCreateUserByEmail(gCtx, email)
	if err != nil {
		c.logger.Error("Failed to get user by email", "error", err)
		gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	err = c.store.UpdateUserVerified(gCtx, email, true)
	if err != nil {
		c.logger.Error("Failed to update user verified", "error", err)
		gCtx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Failed to update user verified"},
		)
		return
	}

	// Generate a new session
	session := sessions.Default(gCtx)
	session.Set("user_id", userID)
	session.Options(sessions.Options{
		Path:   "/",
		MaxAge: 365 * 24 * 60 * 60, // 1 year in secs
	})
	err = session.Save()
	if err != nil {
		c.logger.Error("Failed to save session", "error", err)
		gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Redirect to the homepage
	gCtx.Redirect(http.StatusFound, "/")
}

func (c *Controller) AuthMiddleware() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		session := sessions.Default(gCtx)
		userID := session.Get("user_id")
		if userID == nil {
			gCtx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			gCtx.Abort()
			return
		}
		gCtx.Set("user_id", userID)
		gCtx.Next()
	}
}

func (c *Controller) MeHandler(gCtx *gin.Context) {
	session := sessions.Default(gCtx)
	userID := session.Get("user_id")
	if userID == nil {
		gCtx.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userIDInt, ok := userID.(int64)
	if !ok {
		gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := c.store.GetUserByID(gCtx, userIDInt)
	if err != nil {
		c.logger.Error("Failed to get user by ID", "error", err)
		gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// TODO(pol) return also the lightning address
	gCtx.JSON(http.StatusOK, gin.H{"user": user})
}

func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func (c *Controller) LogoutHandler(gCtx *gin.Context) {
	session := sessions.Default(gCtx)
	user := session.Get("user_id")
	if user == nil {
		gCtx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session token"})
		return
	}
	session.Delete("user_id")

	if err := session.Save(); err != nil {
		gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	gCtx.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func (c *Controller) UpdateProfileHandler(gCtx *gin.Context) {
	session := sessions.Default(gCtx)
	userID := session.Get("user_id")
	if userID == nil {
		gCtx.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userIDInt, ok := userID.(int64)
	if !ok {
		gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		LightningAddress string `json:"lightningAddress"`
	}

	if err := gCtx.ShouldBindJSON(&req); err != nil {
		gCtx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := c.store.UpdateUserLightningAddress(gCtx, userIDInt, req.LightningAddress)
	if err != nil {
		c.logger.Error("Failed to update user lightning address", "error", err)
		gCtx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	gCtx.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func (c *Controller) GetInvoicePreimage(gCtx *gin.Context) {
	paymentHash := gCtx.Param("payment_hash")
	if paymentHash == "" {
		c.logger.Error("payment_hash is required")
		gCtx.JSON(
			http.StatusBadRequest, gin.H{"error": "payment_hash is required"},
		)
		return
	}

	// Check the database first
	status, err := c.store.GetInvoiceStatus(gCtx, paymentHash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.logger.Error("Failed to get invoice status from DB", "error", err)
		gCtx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Failed to check invoice status"},
		)
		return
	}

	now := c.clock.Now()
	// If the invoice is settled or the last check was less than 30 seconds ago,
	// return the cached result
	if status != nil && (status.Settled || now.Sub(status.UpdatedAt) < invoiceCheckInterval) {

		gCtx.JSON(http.StatusOK, gin.H{"status": status})
		return
	}

	// If we reach here, we need to check with the payment provider
	preimage, err := c.invoiceProvider.GetInvoicePreimage(gCtx, paymentHash)
	if err != nil {
		c.logger.Error(
			"Failed to check invoice status with provider", "error", err,
		)
		gCtx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Failed to check invoice status"},
		)
		return
	}

	paid := preimage != ""
	// Update the database with the new status
	status, err = c.store.UpsertInvoiceStatus(gCtx, paymentHash, preimage, paid)
	if err != nil {
		c.logger.Error("Failed to update invoice status in DB", "error", err)
		// We don't return an error to the client here, as we still have the correct status
	}

	gCtx.JSON(http.StatusOK, gin.H{"status": status})
}
