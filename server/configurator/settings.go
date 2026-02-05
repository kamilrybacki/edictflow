package configurator

import (
	"os"
)

type Settings struct {
	DatabaseURL string
	RedisURL    string
	ServerPort  string
	JWTSecret   string
	BaseURL     string
}

func LoadSettings() Settings {
	port := getEnv("SERVER_PORT", "8080")
	return Settings{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/claudeception?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		ServerPort:  port,
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:"+port),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
