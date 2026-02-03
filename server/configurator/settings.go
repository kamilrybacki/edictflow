package configurator

import (
	"os"
)

type Settings struct {
	DatabaseURL string
	ServerPort  string
	JWTSecret   string
}

func LoadSettings() Settings {
	return Settings{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/claudeception?sslmode=disable"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
