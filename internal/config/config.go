package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL  string
	Port         string
	Timeout      time.Duration
	MaxRedirects int
	MaxBodyBytes int64
	UserAgent    string
	Interval     time.Duration
	BatchSize    int
	Concurrency  int
	Lease        time.Duration
	HostInterval time.Duration
}

func LoadConfig() *Config {
	return &Config{
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://localhost:5432/mydb"),
		Port:         getEnv("PORT", "8080"),
		Timeout:      getEnvDuration("TIMEOUT", 10*time.Second),
		MaxRedirects: getEnvInt("MAX_REDIRECTS", 5),
		MaxBodyBytes: getEnvInt64("MAX_BODY_BYTES", 16<<10),
		UserAgent:    getEnv("USER_AGENT", "LinkOrDie/1.0"),
		Interval:     getEnvDuration("INTERVAL", 30*time.Second),
		BatchSize:    getEnvInt("BATCH_SIZE", 50),
		Concurrency:  getEnvInt("CONCURRENCY", 8),
		Lease:        getEnvDuration("LEASE", 5*time.Minute),
		HostInterval: getEnvDuration("HOST_INTERVAL", 2*time.Second),
	}
}

// Helper functions for safe parsing with fallbacks

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(valStr)
	if err != nil {
		return defaultVal
	}
	return d
}

func getEnvInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return i
}

func getEnvInt64(key string, defaultVal int64) int64 {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	i, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return defaultVal
	}
	return i
}
