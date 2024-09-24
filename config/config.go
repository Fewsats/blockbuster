package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/fewsats/blockbuster/auth"
	"github.com/fewsats/blockbuster/cloudflare"
	"github.com/fewsats/blockbuster/email"
	"github.com/fewsats/blockbuster/lightning"
	"github.com/fewsats/blockbuster/store"
	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
)

const (
	DefaultLogLevel = "info"
	DefaultPort     = 8080
)

type EmailConfig struct {
	Provider string `long:"provider" description:"Email provider"`
	APIKey   string `long:"api_key" description:"Email provider API key"`
}

type StorageConfig struct {
	Provider string `long:"provider" description:"Storage provider"`
	Local    struct {
		Path string `long:"path" description:"Local storage path"`
	} `group:"local" namespace:"local"`
}

type Config struct {
	LogLevel string `long:"log_level" description:"Logging level {debug, info, warn, error}"`
	Port     int    `long:"port" description:"Port to listen on"`
	GinMode  string `long:"gin_mode" description:"Gin mode {debug, release}"`

	Auth       auth.Config       `group:"auth" namespace:"auth"`
	Email      email.Config      `group:"email" namespace:"email"`
	Cloudflare cloudflare.Config `group:"cloudflare" namespace:"cloudflare"`
	Store      store.Config      `group:"store" namespace:"store"`
	Lightning  lightning.Config  `group:"lightning" namespace:"lightning"`
}

func (c *Config) Validate() error {
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	return nil
}

func (c *Config) SetLoggerLevel() error {
	switch c.LogLevel {
	case "info":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "debug":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "warn":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "error":
		slog.SetLogLoggerLevel(slog.LevelError)
	default:
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}
	return nil
}

func (c *Config) SetGinMode() {
	if c.GinMode != "" {
		gin.SetMode(c.GinMode)
	}
}

func DefaultConfig() *Config {
	return &Config{
		LogLevel: DefaultLogLevel,
		Port:     DefaultPort,
		GinMode:  gin.DebugMode,

		Auth:       *auth.DefaultConfig(),
		Email:      *email.DefaultConfig(),
		Store:      *store.DefaultConfig(),
		Cloudflare: *cloudflare.DefaultConfig(),
		Lightning:  *lightning.DefaultConfig(),
	}
}

func LoadConfig(logger *slog.Logger) (*Config, error) {
	cfg := DefaultConfig()
	if _, err := flags.Parse(cfg); err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}

	configFilePath := os.Getenv("BLOCKBUSTER_CONFIG")
	if configFilePath == "" {
		configFilePath = "blockbuster.conf"
	}

	logger.Info(
		"Configuration file",
		"path", configFilePath,
	)

	fileParser := flags.NewParser(cfg, flags.Default)
	err := flags.NewIniParser(fileParser).ParseFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config file: %w", err)
	}

	flagParser := flags.NewParser(cfg, flags.Default)
	if _, err := flagParser.Parse(); err != nil {
		return nil, err
	}

	return cfg, nil
}
