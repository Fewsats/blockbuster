package lightning

const (
	// ProviderAlby is the Alby provider.
	ProviderAlby = "alby"
)

type AlbyConfig struct {
	// APIKey is the API key to use for the Alby provider.
	APIKey string `long:"api_key" description:"API key for the Alby."`
}

// Config is the main config for the lightning service.
type Config struct {
	// Provider is the provider to use for creating lightning invoices.
	Provider string `long:"provider" description:"LN provider to use."`

	// AlbyConfig is Alby's configuration.
	Alby AlbyConfig `group:"alby" namespace:"alby"`
}

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() *Config {
	return &Config{
		Provider: ProviderAlby,
	}
}
