package email

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() *Config {
	return &Config{
		Provider: "resend",
		APIKey:   "",
	}
}

// Config holds the email configuration.
type Config struct {
	Provider string `long:"provider" description:"Email provider"`
	APIKey   string `long:"api_key" description:"Email provider API key"`
}
