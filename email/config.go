package email

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() *Config {
	return &Config{
		Provider: "resend",
		APIKey:   "",
		BaseURL:  "http://localhost:8080",
	}
}

// Config holds the email configuration.
type Config struct {
	Provider string `long:"provider" description:"Email provider"`
	APIKey   string `long:"api_key" description:"Email provider API key"`
	BaseURL  string `long:"base_url" description:"Base URL for the application"`
}
