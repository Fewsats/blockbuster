package main

import (
	"log/slog"
	"os"

	"github.com/fewsats/blockbuster/auth"
	"github.com/fewsats/blockbuster/cloudflare"
	"github.com/fewsats/blockbuster/config"
	"github.com/fewsats/blockbuster/email"
	"github.com/fewsats/blockbuster/server"
	storePkg "github.com/fewsats/blockbuster/store"
	"github.com/fewsats/blockbuster/utils"
	"github.com/fewsats/blockbuster/video"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	emailService := email.NewResendService(logger, &cfg.Email)
	cloudflareService, err := cloudflare.NewService(&cfg.Cloudflare)
	if err != nil {
		logger.Error("Failed to create cloudflare service", "error", err)
		return
	}
	// Initialize the store.
	clock := utils.NewRealClock()
	store, err := storePkg.NewStore(logger, &cfg.Store, clock)
	if err != nil {
		logger.Error("Failed to create store", "error", err)
		os.Exit(1)
	}

	authController := auth.NewController(emailService, logger, store, &cfg.Auth)
	videoController := video.NewController(cloudflareService, store, logger)

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
