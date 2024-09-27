package video

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() *Config {
	return &Config{
		L402BaseURL: "http://localhost:8080/video/stream",
		L402InfoURL: "http://localhost:8080/video/info",
	}
}

type Config struct {
	L402BaseURL string `long:"l402_base_url" description:"L402 base URL"`
	L402InfoURL string `long:"l402_info_url" description:"L402 info URL"`
}
