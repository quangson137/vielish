package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	AppEnv      string
	CORSOrigins []string
}

func Load() (Config, error) {
	appEnv := getEnv("APP_ENV", "development")
	jwtSecret := getEnv("JWT_SECRET", "")

	// In production, JWT_SECRET is required
	if appEnv == "production" && jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required in production")
	}

	// In development, use a default secret
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production"
	}

	corsOriginsStr := getEnv("CORS_ORIGINS", "http://localhost:3000")
	corsOrigins := strings.Split(corsOriginsStr, ",")
	for i := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
	}

	return Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:   jwtSecret,
		AppEnv:      appEnv,
		CORSOrigins: corsOrigins,
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
