package store

import (
	"fmt"
)

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() *Config {
	return &Config{
		SkipMigrations:     false,
		ConnectionString:   "postgres://postgres-dev:postgres-dev@localhost:5432/blockbuster-dev?sslmode=disable",
		MaxOpenConnections: 25,
		MigrationsPath:     "store/migrations",
	}
}

// Config holds the database configuration.
type Config struct {
	SkipMigrations     bool   `long:"skip_migrations" description:"Skip applying migrations on startup."`
	ConnectionString   string `long:"connection_string" description:"Database connection string."`
	MaxOpenConnections int32  `long:"max_connections" description:"Max open connections to keep alive to the database server."`
	MigrationsPath     string `long:"migrations_path" description:"Path to the migrations folder"`
}

// DSN returns the connection string to connect to the database.
func (s *Config) DSN(hidePassword bool) string {
	if !hidePassword {
		return s.ConnectionString
	}

	// Simple password hiding for logging purposes
	return fmt.Sprintf("%s****%s",
		s.ConnectionString[:15],
		s.ConnectionString[len(s.ConnectionString)-20:])
}
