package infrastructure

import (
	"os"
	"strconv"
)

// DatabaseConfigFromEnv creates a database configuration from environment variables
func DatabaseConfigFromEnv() DatabaseConfig {
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite3" // Default to SQLite
	}

	switch driver {
	case "postgres", "postgresql":
		return DatabaseConfig{
			Driver:   "postgres",
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvIntOrDefault("DB_PORT", 5432),
			Database: getEnvOrDefault("DB_NAME", "goro_ldp"),
			Username: getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", ""),
			SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
		}
	default: // SQLite
		return DatabaseConfig{
			Driver:   "sqlite3",
			Database: getEnvOrDefault("DB_PATH", "./data/membership.db"),
		}
	}
}

// Example configurations for different environments

// DevelopmentSQLiteConfig returns a SQLite configuration for development
func DevelopmentSQLiteConfig() DatabaseConfig {
	return DefaultSQLiteConfig("./data/dev_membership.db")
}

// TestSQLiteConfig returns a SQLite configuration for testing
func TestSQLiteConfig() DatabaseConfig {
	return DefaultSQLiteConfig(":memory:")
}

// ProductionPostgreSQLConfig returns a PostgreSQL configuration for production
func ProductionPostgreSQLConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver:   "postgres",
		Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:     getEnvIntOrDefault("POSTGRES_PORT", 5432),
		Database: getEnvOrDefault("POSTGRES_DB", "goro_ldp_prod"),
		Username: getEnvOrDefault("POSTGRES_USER", "goro_user"),
		Password: os.Getenv("POSTGRES_PASSWORD"), // Required in production
		SSLMode:  getEnvOrDefault("POSTGRES_SSLMODE", "require"),
	}
}

// StagingPostgreSQLConfig returns a PostgreSQL configuration for staging
func StagingPostgreSQLConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver:   "postgres",
		Host:     getEnvOrDefault("POSTGRES_HOST", "staging-db.example.com"),
		Port:     getEnvIntOrDefault("POSTGRES_PORT", 5432),
		Database: getEnvOrDefault("POSTGRES_DB", "goro_ldp_staging"),
		Username: getEnvOrDefault("POSTGRES_USER", "goro_staging"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		SSLMode:  getEnvOrDefault("POSTGRES_SSLMODE", "require"),
	}
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
