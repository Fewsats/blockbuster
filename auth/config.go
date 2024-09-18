package auth

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() *Config {
	return &Config{
		TokenExpirationMinutes: 15,
		SessionSecret:          "super-secret-string",
	}
}

type Config struct {
	TokenExpirationMinutes int    `long:"token_expiration_minutes" description:"Token expiration duration in minutes"`
	SessionSecret          string `long:"session_secret" description:"Session secret"`
}
