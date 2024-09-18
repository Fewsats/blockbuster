package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                int
	DatabasePath        string
	ResendAPIKey        string
	BaseURL             string
	JWTSecret           string
	VideoUploadPath     string
	ThumbnailUploadPath string
	SessionSecret       string
}

func LoadConfig() (*Config, error) {
	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:                port,
		DatabasePath:        getEnv("DATABASE_PATH", "blockbuster.db"),
		ResendAPIKey:        getEnv("RESEND_API_KEY", ""),
		BaseURL:             getEnv("BASE_URL", "http://localhost:8080"),
		JWTSecret:           getEnv("JWT_SECRET", "your-secret-key"),
		VideoUploadPath:     getEnv("VIDEO_UPLOAD_PATH", "./uploads/videos"),
		ThumbnailUploadPath: getEnv("THUMBNAIL_UPLOAD_PATH", "./uploads/thumbnails"),
		SessionSecret:       getEnv("SESSION_SECRET", "your-secret-key"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
