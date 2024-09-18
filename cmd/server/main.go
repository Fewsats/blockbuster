package main

import (
	"log/slog"
	"os"

	"github.com/Fewsats/blockbuster/auth"
	"github.com/Fewsats/blockbuster/config"
	"github.com/Fewsats/blockbuster/database"
	"github.com/Fewsats/blockbuster/email"
	"github.com/Fewsats/blockbuster/server"
	"github.com/Fewsats/blockbuster/video"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := database.NewSQLiteDB(cfg.DatabasePath)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	emailService := email.NewResendService(cfg.ResendAPIKey)
	authController := auth.NewController(logger, db, emailService, cfg.BaseURL)

	videoController := video.NewController(db, cfg.VideoUploadPath, cfg.ThumbnailUploadPath)

	srv, err := server.NewServer(logger, cfg, authController, videoController)
	if err != nil {
		logger.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	logger.Info("Starting server", "port", cfg.Port)
	if err := srv.Run(); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}
}
