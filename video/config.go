package video

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() *Config {
	return &Config{
		L402BaseURL: "http://localhost:8080/video/stream",
	}
}

type Config struct {
	L402BaseURL string `long:"l402_base_url" description:"L402 base URL"`
}
