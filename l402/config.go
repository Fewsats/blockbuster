package l402

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() *Config {
	return &Config{
		Domain: "localhost:8080",
	}
}

type Config struct {
	Domain string `long:"domain" description:"Domain"`
}
