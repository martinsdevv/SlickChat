package config

type MediaConfig struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	Bucket          string
	UseSSL          bool
	PublicBaseURL   string
	PresignUploadTTL int // seconds
	PresignReadTTL   int // seconds
}

func LoadMediaConfig() MediaConfig {
	return MediaConfig{
		Endpoint:         getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey:        getEnv("MINIO_ACCESS_KEY", "slickchat"),
		SecretKey:        getEnv("MINIO_SECRET_KEY", "slickchat-secret"),
		Bucket:           getEnv("MINIO_BUCKET", "slickchat-media"),
		UseSSL:           getEnv("MINIO_USE_SSL", "false") == "true",
		PublicBaseURL:    getEnv("MINIO_PUBLIC_URL", "http://localhost:9000"),
		PresignUploadTTL: 900,
		PresignReadTTL:   3600,
	}
}

// UseProxyUpload sends uploads through the API (required for Cloudflare Tunnel / remote clients).
// Set MINIO_UPLOAD_DIRECT=true only for local dev without the demo proxy.
func UseProxyUpload() bool {
	return getEnv("MINIO_UPLOAD_DIRECT", "false") != "true"
}
