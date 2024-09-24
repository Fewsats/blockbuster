package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/fewsats/blockbuster/auth"
	"github.com/fewsats/blockbuster/cloudflare"
	"github.com/fewsats/blockbuster/config"
	"github.com/fewsats/blockbuster/email"
	"github.com/fewsats/blockbuster/l402"
	"github.com/fewsats/blockbuster/lightning"
	"github.com/fewsats/blockbuster/orders"
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

	if err := cfg.SetLoggerLevel(); err != nil {
		logger.Error(
			"Unable to set logger level",
			"error", err,
		)

		return
	}

	cfg.SetGinMode()

	// Initialize the store.
	clock := utils.NewRealClock()
	store, err := storePkg.NewStore(logger, &cfg.Store, clock)
	if err != nil {
		logger.Error("Failed to create store", "error", err)
		os.Exit(1)
	}

	emailService := email.NewResendService(logger, &cfg.Email)
	cloudflareService, err := cloudflare.NewService(&cfg.Cloudflare)
	if err != nil {
		logger.Error("Failed to create cloudflare service", "error", err)
		return
	}

	var invoiceProvider l402.InvoiceProvider
	switch cfg.Lightning.Provider {
	case lightning.ProviderAlby:
		invoiceProvider = lightning.NewAlbyProvider(
			http.DefaultClient, cfg.Lightning.Alby.APIKey,
		)

	default:
		logger.Error(
			"Unknown lightning provider",
			"provider", cfg.Lightning.Provider,
		)
	}

	authenticator := l402.NewAuthenticator(
		logger, invoiceProvider, store, clock,
	)

	// Managers
	ordersMgr := orders.NewManager(logger, store)

	authController := auth.NewController(emailService, logger, store, &cfg.Auth)
	videoController := video.NewController(cloudflareService, ordersMgr,
		authenticator, store, logger, clock)

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
