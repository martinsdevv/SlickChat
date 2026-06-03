package media

import (
	"strings"

	"github.com/martinsdevv/slickchat/infrastructure/config"
)

// ShouldUseProxyUpload is true when uploads must go through the API (default for demo/tunnel).
func ShouldUseProxyUpload(publicBaseURL string) bool {
	if config.UseProxyUpload() {
		return true
	}
	base := strings.TrimSpace(strings.ToLower(publicBaseURL))
	if base == "" {
		return false
	}
	return !strings.Contains(base, "localhost") && !strings.Contains(base, "127.0.0.1")
}
