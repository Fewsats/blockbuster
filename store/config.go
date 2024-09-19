package store

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() *Config {
	return &Config{
		SkipMigrations:     false,
		ConnectionString:   "blockbuster.db",
		MaxOpenConnections: 25,
		MigrationsPath:     "store/sqlc/migrations",
	}
}

// Config holds the database configuration.
type Config struct {
	SkipMigrations     bool   `long:"skip_migrations" description:"Skip applying migrations on startup."`
	ConnectionString   string `long:"connection_string" description:"Database connection string."`
	MaxOpenConnections int32  `long:"max_connections" description:"Max open connections to keep alive to the database server."`
	MigrationsPath     string `long:"migrations_path" description:"Path to the migrations folder"`
}
