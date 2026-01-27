package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Worker   WorkerConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port string
	Env  string // development, staging, production
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// WorkerConfig holds Asynq worker configuration
type WorkerConfig struct {
	Concurrency int // Number of concurrent workers
	// RetentionMinutes controls how long to keep completed/failed task records in Redis
	// so they show up in Asynqmon (0 = do not keep).
	RetentionMinutes int
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "admin"),
			Password: getEnv("DB_PASSWORD", "secret123"),
			DBName:   getEnv("DB_NAME", "taskqueue"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Worker: WorkerConfig{
			Concurrency: getEnvAsInt("WORKER_CONCURRENCY", 20),
			RetentionMinutes: getEnvAsInt("ASYNQ_RETENTION_MINUTES", 30),
		},
	}

	return cfg, nil
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt reads an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// GetDatabaseDSN returns PostgreSQL connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// GetRedisAddr returns Redis address
func (c *Config) GetRedisAddr() string {
	return c.Redis.Addr
}
