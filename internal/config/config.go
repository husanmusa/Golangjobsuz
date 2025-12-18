package config

import "os"

// Config holds environment driven configuration for the API server.
type Config struct {
	DatabaseURL string
	Port        string
	AIEndpoint  string
}

// FromEnv loads configuration with defaults when values are missing.
func FromEnv() (*Config, error) {
	dbURL := getenv("DATABASE_URL", "")
	port := getenv("PORT", "8080")
	aiEndpoint := getenv("AI_ENDPOINT", "http://localhost:18081/complete")

	return &Config{DatabaseURL: dbURL, Port: port, AIEndpoint: aiEndpoint}, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
