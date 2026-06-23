package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort       string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	WorkerPoolSize   int
	WorkerBufferSize int
}

func Load() (*Config, error) {
	_ = godotenv.Load("config.env")

	poolSize, err := getEnvAsInt("WORKER_POOL_SIZE", 100)
	if err != nil {
		return nil, fmt.Errorf("invalid WORKER_POOL_SIZE: %w", err)
	}

	bufSize, err := getEnvAsInt("WORKER_BUFFER_SIZE", 500)
	if err != nil {
		return nil, fmt.Errorf("invalid WORKER_BUFFER_SIZE: %w", err)
	}

	return &Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", ""),
		DBName:           getEnv("DB_NAME", "wallet"),
		DBSSLMode:        getEnv("DB_SSLMODE", "disable"),
		WorkerPoolSize:   poolSize,
		WorkerBufferSize: bufSize,
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int) (int, error) {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s as int: %w", key, err)
	}

	return value, nil
}
