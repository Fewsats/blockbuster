package cloudflare

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() *Config {
	return &Config{}
}

type Config struct {
	Endpoint         *string `long:"endpoint" description:"Cloudflare R2 endpoint."`
	AccessKey        string  `long:"access_key" description:"Cloudflare R2 API token."`
	SecretAccessKey  string  `long:"secret_access_key" description:"Cloudflare R2 API token."`
	PublicBucketName string  `long:"public_bucket_name" description:"Cloudflare R2 public bucket name."`
	APIToken         string  `long:"api_token" description:"Cloudflare API token for Streams"`
	AccountID        string  `long:"account_id" description:"Cloudflare account ID"`
}
