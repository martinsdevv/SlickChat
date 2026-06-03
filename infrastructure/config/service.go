package config

// KafkaBroker returns the Kafka bootstrap address for app services.
func KafkaBroker() string {
	return getEnv("KAFKA_BROKERS", "localhost:9092")
}

// RedisAddr returns the Redis address (host:port).
func RedisAddr() string {
	return getEnv("REDIS_ADDR", "localhost:6379")
}

// APIInternalURL is the HTTP base URL of the API (used by the gateway).
func APIInternalURL() string {
	return getEnv("SLICKCHAT_API_URL", "http://127.0.0.1:8081")
}
