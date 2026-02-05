package configurator

import (
	"os"
)

type Settings struct {
	DatabaseURL         string
	RedisURL            string
	ServerPort          string
	JWTSecret           string
	BaseURL             string
	SplunkEnabled       bool
	SplunkHECURL        string
	SplunkHECToken      string
	SplunkSource        string
	SplunkSourceType    string
	SplunkIndex         string
	SplunkSkipTLSVerify bool
}

func LoadSettings() Settings {
	port := getEnv("SERVER_PORT", "8080")
	return Settings{
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://localhost:5432/claudeception?sslmode=disable"),
		RedisURL:            getEnv("REDIS_URL", "redis://localhost:6379/0"),
		ServerPort:          port,
		JWTSecret:           getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		BaseURL:             getEnv("BASE_URL", "http://localhost:"+port),
		SplunkEnabled:       getEnv("SPLUNK_ENABLED", "false") == "true",
		SplunkHECURL:        getEnv("SPLUNK_HEC_URL", ""),
		SplunkHECToken:      getEnv("SPLUNK_HEC_TOKEN", ""),
		SplunkSource:        getEnv("SPLUNK_SOURCE", "claudeception"),
		SplunkSourceType:    getEnv("SPLUNK_SOURCETYPE", "claudeception:metrics"),
		SplunkIndex:         getEnv("SPLUNK_INDEX", "main"),
		SplunkSkipTLSVerify: getEnv("SPLUNK_SKIP_TLS_VERIFY", "false") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
