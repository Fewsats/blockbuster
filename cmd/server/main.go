package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/fewsats/blockbuster/auth"
	"github.com/fewsats/blockbuster/config"
	"github.com/fewsats/blockbuster/database"
	"github.com/fewsats/blockbuster/email"
	"github.com/fewsats/blockbuster/server"
	"github.com/fewsats/blockbuster/video"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := database.NewSQLiteDB(cfg.Store.ConnectionString)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	emailService := email.NewResendService(logger, &cfg.Email)
	authController := auth.NewController(emailService, logger, db, &cfg.Auth)

	// TODO(pol) use cloudflare or external service
	videoController := video.NewController(db,
		fmt.Sprintf("%s/%s", cfg.Storage.Local.Path, "videos"),
		fmt.Sprintf("%s/%s", cfg.Storage.Local.Path, "thumbnails"),
	)

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
