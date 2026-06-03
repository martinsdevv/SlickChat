package media

import (
	"net/url"
	"strings"
)

// PublicizePresignedURL rewrites a MinIO presigned URL for external clients (ex.: Cloudflare Tunnel).
// publicBase should be like https://your-host.trycloudflare.com/storage (no trailing slash).
func PublicizePresignedURL(rawURL, publicBase string) string {
	publicBase = strings.TrimRight(strings.TrimSpace(publicBase), "/")
	if publicBase == "" {
		return rawURL
	}

	raw, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	pub, err := url.Parse(publicBase)
	if err != nil {
		return rawURL
	}

	if raw.Scheme == pub.Scheme && raw.Host == pub.Host && (pub.Path == "" || pub.Path == "/") {
		return rawURL
	}

	raw.Scheme = pub.Scheme
	raw.Host = pub.Host
	if pub.Path != "" && pub.Path != "/" {
		raw.Path = strings.TrimSuffix(pub.Path, "/") + raw.Path
	}

	return raw.String()
}
