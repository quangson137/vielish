package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear env vars that might interfere
	os.Unsetenv("APP_ENV")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("REDIS_URL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.DatabaseURL != "postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable" {
		t.Errorf("DatabaseURL = %q, want default", cfg.DatabaseURL)
	}
	if cfg.RedisURL != "redis://localhost:6379" {
		t.Errorf("RedisURL = %q, want default", cfg.RedisURL)
	}
	if cfg.JWTSecret != "dev-secret-change-in-production" {
		t.Errorf("JWTSecret = %q, want dev default", cfg.JWTSecret)
	}
	if cfg.AppEnv != "development" {
		t.Errorf("AppEnv = %q, want %q", cfg.AppEnv, "development")
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("PORT", "3000")
	os.Setenv("DATABASE_URL", "postgres://custom:custom@db:5432/custom?sslmode=disable")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("DATABASE_URL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != "3000" {
		t.Errorf("Port = %q, want %q", cfg.Port, "3000")
	}
	if cfg.DatabaseURL != "postgres://custom:custom@db:5432/custom?sslmode=disable" {
		t.Errorf("DatabaseURL = %q, want custom", cfg.DatabaseURL)
	}
}

func TestLoad_Production_RequiresJWTSecret(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("APP_ENV")

	_, err := Load()
	if err == nil {
		t.Error("Load() should fail in production without JWT_SECRET")
	}
}

func TestLoad_Production_WithJWTSecret(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Setenv("JWT_SECRET", "my-prod-secret")
	defer os.Unsetenv("APP_ENV")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.JWTSecret != "my-prod-secret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "my-prod-secret")
	}
}
