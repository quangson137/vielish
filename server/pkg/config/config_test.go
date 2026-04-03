package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sonpham/vielish/server/pkg/config"
)

const testYAML = `
app:
  port: "9090"
  env: test
database:
  url: postgres://test:test@localhost:5432/testdb
redis:
  url: redis://localhost:6379
jwt:
  secret: test-secret
  access_ttl: 2h
  refresh_ttl: 336h
cors:
  origins:
    - http://localhost:3000
tracing:
  enabled: false
  endpoint: localhost:4317
`

func TestLoad_ReadsYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(testYAML), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Port != "9090" {
		t.Errorf("App.Port = %q, want %q", cfg.App.Port, "9090")
	}
	if cfg.JWT.AccessTTL != 2*time.Hour {
		t.Errorf("JWT.AccessTTL = %v, want 2h", cfg.JWT.AccessTTL)
	}
	if cfg.JWT.RefreshTTL != 336*time.Hour {
		t.Errorf("JWT.RefreshTTL = %v, want 336h", cfg.JWT.RefreshTTL)
	}
	if len(cfg.CORS.Origins) != 1 || cfg.CORS.Origins[0] != "http://localhost:3000" {
		t.Errorf("CORS.Origins = %v, want [http://localhost:3000]", cfg.CORS.Origins)
	}
}

func TestLoad_EnvOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(testYAML), 0644)
	t.Setenv("APP_PORT", "7777")

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.App.Port != "7777" {
		t.Errorf("App.Port = %q after env override, want %q", cfg.App.Port, "7777")
	}
}
