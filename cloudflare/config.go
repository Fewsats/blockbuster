package cloudflare

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() *Config {
	return &Config{}
}

type Config struct {
	Endpoint         *string `long:"endpoint" description:"Cloudflare R2 endpoint."`
	AccessKey        string  `long:"access_key" description:"Cloudflare API token."`
	SecretAccessKey  string  `long:"secret_access_key" description:"Cloudflare API token."`
	VideoBucketName  string  `long:"video_bucket_name" description:"Cloudflare R2 private video bucket name."`
	PublicBucketName string  `long:"public_bucket_name" description:"Cloudflare R2 public bucket name."`
}
