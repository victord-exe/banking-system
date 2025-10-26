package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	ServerPort string
	GinMode    string

	// PostgreSQL configuration
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresDSN      string

	// TigerBeetle configuration
	TigerBeetleHost    string
	TigerBeetlePort    string
	TigerBeetleAddress string // Full address (host:port)

	// JWT configuration
	JWTSecret string

	// OpenRouter/AI configuration
	OpenRouterAPIKey string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (ignore error in production)
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		GinMode:    getEnv("GIN_MODE", "release"),

		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "banking_user"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "banking_secure_password_2024"),
		PostgresDB:       getEnv("POSTGRES_DB", "banking_system"),

		TigerBeetleHost: getEnv("TIGERBEETLE_HOST", "localhost"),
		TigerBeetlePort: getEnv("TIGERBEETLE_PORT", "3000"),

		JWTSecret:        getEnv("JWT_SECRET", ""),
		OpenRouterAPIKey: getEnv("OPENROUTER_API_KEY", ""),
	}

	// Build PostgreSQL DSN if not provided
	cfg.PostgresDSN = getEnv("POSTGRES_DSN", "")
	if cfg.PostgresDSN == "" {
		cfg.PostgresDSN = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.PostgresHost,
			cfg.PostgresPort,
			cfg.PostgresUser,
			cfg.PostgresPassword,
			cfg.PostgresDB,
		)
	}

	// Build TigerBeetle address - allow override via TIGERBEETLE_ADDRESS env var
	cfg.TigerBeetleAddress = getEnv("TIGERBEETLE_ADDRESS", "")
	if cfg.TigerBeetleAddress == "" {
		cfg.TigerBeetleAddress = fmt.Sprintf("%s:%s", cfg.TigerBeetleHost, cfg.TigerBeetlePort)
	}

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks that all required configuration is present
func (c *Config) validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	return nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
