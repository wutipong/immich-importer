package immich

import "net/url"

type ServerConfig struct {
	URL    *url.URL
	APIKey string
	DryRun bool
}
