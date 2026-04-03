# Backend Refactor: DDD/Clean Architecture Implementation Plan

> Steps use checkbox (`- [ ]`) syntax for tracking progress.

**Goal:** Refactor `server/` from flat per-feature packages to DDD/Clean Architecture wired with Uber fx, using GORM, Viper, and Zap.

**Architecture:** Four layers — `domain/` (entities + interfaces, no framework deps) → `appcore/` (use cases + DTOs) → `driven/` (GORM/Redis adapters) + `driving/` (Gin HTTP adapters). All components injected via Uber fx modules. Big-bang refactor: old `internal/auth/`, `internal/config/`, `internal/database/`, `internal/router/`, `pkg/response/` are deleted at the end.

**Tech Stack:** Go, Gin, `go.uber.org/fx`, `gorm.io/gorm` + `gorm.io/driver/postgres`, `github.com/spf13/viper`, `go.uber.org/zap`, OpenTelemetry SDK + OTLP exporter, `github.com/redis/go-redis/v9`.

---

## Package Name Conventions

| Path | `package` name | Imported as (alias when needed) |
|---|---|---|
| `internal/domain/auth/` | `domain` | `authdom` when multiple auth pkgs in scope |
| `internal/appcore/auth/` | `appcore` | `authcore` |
| `internal/driven/auth/` | `driven` | `authdriven` |
| `internal/driven/database/` | `database` | — |
| `internal/driving/httpui/` | `httpui` | — |
| `internal/driving/httpui/handler/` | `handler` | — |
| `internal/driving/httpui/presenter/` | `presenter` | — |
| `internal/driving/httpui/middleware/` | `middleware` | — |
| `pkg/config/` | `config` | — |
| `pkg/log/` | `pkglog` (declared as `log`) | `pkglog` |
| `pkg/ctxbase/` | `ctxbase` | — |
| `pkg/httpbase/` | `httpbase` | — |
| `pkg/tracing/` | `tracing` | — |

> **Design deviation from spec:** `UseCase.NewUseCase` takes `config.Config` (in addition to repo + service) to read `JWT.AccessTTL` and `JWT.RefreshTTL`. The spec omits this param, but the `StoreRefreshToken(ttl)` call requires it.

---

## File Map

**Create:**
```
server/
├── cmd/api/
│   ├── main.go                                        (rewrite)
│   └── config.yaml                                    (new)
├── pkg/
│   ├── config/config.go                               (new)
│   ├── config/config_test.go                          (new)
│   ├── ctxbase/ctxbase.go                             (new)
│   ├── ctxbase/ctxbase_test.go                        (new)
│   ├── httpbase/httpbase.go                           (new)
│   ├── httpbase/httpbase_test.go                      (new)
│   ├── log/log.go                                     (new)
│   └── tracing/tracing.go                             (new)
├── internal/
│   ├── driven/database/gorm.go                        (new)
│   ├── driven/database/redis.go                       (new)
│   ├── driven/database/module.go                      (new)
│   ├── domain/auth/entity.go                          (new)
│   ├── domain/auth/errors.go                          (new)
│   ├── domain/auth/repository.go                      (new)
│   ├── domain/auth/service.go                         (new)
│   ├── domain/auth/service_test.go                    (new)
│   ├── domain/auth/module.go                          (new)
│   ├── appcore/auth/dto.go                            (new)
│   ├── appcore/auth/usecase.go                        (new)
│   ├── appcore/auth/usecase_test.go                   (new)
│   ├── appcore/auth/module.go                         (new)
│   ├── driven/auth/gorm_model.go                      (new)
│   ├── driven/auth/repository.go                      (new)
│   ├── driven/auth/module.go                          (new)
│   ├── driving/httpui/handler/auth_handler.go         (new)
│   ├── driving/httpui/handler/auth_handler_test.go    (new)
│   ├── driving/httpui/presenter/auth_presenter.go     (new)
│   ├── driving/httpui/middleware/auth.go              (new)
│   ├── driving/httpui/middleware/auth_test.go         (new)
│   ├── driving/httpui/server.go                       (new)
│   └── driving/httpui/module.go                       (new)
```

**Delete (Task 13):**
```
server/internal/auth/          (all files)
server/internal/config/        (all files)
server/internal/database/      (all files)
server/internal/router/        (all files)
server/pkg/response/           (all files)
```

---

## Task 1: Add New Dependencies

**Files:**
- Modify: `server/go.mod` (via `go get`)

- [x] **Step 1: Add fx, GORM, Viper, Zap, OpenTelemetry**

Run from `server/`:
```bash
go get go.uber.org/fx@latest
go get gorm.io/gorm@latest
go get gorm.io/driver/postgres@latest
go get github.com/spf13/viper@latest
go get go.uber.org/zap@latest
go get go.opentelemetry.io/otel@latest
go get go.opentelemetry.io/otel/sdk@latest
go get go.opentelemetry.io/otel/sdk/trace@latest
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@latest
go mod tidy
```

Expected: `go.mod` now lists all new direct dependencies, `go.sum` updated. `go mod tidy` exits 0.

---

## Task 2: pkg/config — Viper Config Loader

**Files:**
- Create: `server/pkg/config/config.go`
- Create: `server/pkg/config/config_test.go`

- [x] **Step 1: Write the failing test**

`server/pkg/config/config_test.go`:
```go
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
```

- [x] **Step 2: Run to verify it fails**

```bash
cd server && go test ./pkg/config/... -run TestLoad -v
```

Expected: `FAIL` — `package config` undefined.

- [x] **Step 3: Implement pkg/config/config.go**

```go
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Tracing  TracingConfig
}

type AppConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	URL string
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}

type CORSConfig struct {
	Origins []string
}

type TracingConfig struct {
	Enabled  bool
	Endpoint string
}

// Load reads config.yaml from the given search paths (or "." and "./cmd/api" by default)
// and applies environment variable overrides. Env vars use underscore-separated keys,
// e.g. APP_PORT overrides app.port, JWT_SECRET overrides jwt.secret.
func Load(searchPaths ...string) (Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if len(searchPaths) == 0 {
		v.AddConfigPath(".")
		v.AddConfigPath("./cmd/api")
	} else {
		for _, p := range searchPaths {
			v.AddConfigPath(p)
		}
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, fmt.Errorf("reading config file: %w", err)
		}
	}

	var cfg Config
	decodeHook := func(dc *mapstructure.DecoderConfig) {
		dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			dc.DecodeHook,
		)
	}
	if err := v.Unmarshal(&cfg, decodeHook); err != nil {
		return Config{}, fmt.Errorf("unmarshalling config: %w", err)
	}
	return cfg, nil
}

// NewConfig is the fx provider — loads from default search paths.
func NewConfig() (Config, error) {
	return Load()
}

var Module = fx.Module("config",
	fx.Provide(NewConfig),
)
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd server && go test ./pkg/config/... -run TestLoad -v
```

Expected: `PASS` for both `TestLoad_ReadsYAML` and `TestLoad_EnvOverridesYAML`.

---

## Task 3: pkg/ctxbase — Typed Context Keys

**Files:**
- Create: `server/pkg/ctxbase/ctxbase.go`
- Create: `server/pkg/ctxbase/ctxbase_test.go`

- [x] **Step 1: Write the failing test**

`server/pkg/ctxbase/ctxbase_test.go`:
```go
package ctxbase_test

import (
	"context"
	"testing"

	"github.com/sonpham/vielish/server/pkg/ctxbase"
)

func TestSetGetUserID(t *testing.T) {
	ctx := ctxbase.SetUserID(context.Background(), "user-abc")
	id, ok := ctxbase.GetUserID(ctx)
	if !ok {
		t.Fatal("GetUserID returned ok=false")
	}
	if id != "user-abc" {
		t.Errorf("GetUserID = %q, want %q", id, "user-abc")
	}
}

func TestGetUserID_Missing(t *testing.T) {
	_, ok := ctxbase.GetUserID(context.Background())
	if ok {
		t.Error("GetUserID should return ok=false on empty context")
	}
}

func TestSetGetRequestID(t *testing.T) {
	ctx := ctxbase.SetRequestID(context.Background(), "req-123")
	id, ok := ctxbase.GetRequestID(ctx)
	if !ok {
		t.Fatal("GetRequestID returned ok=false")
	}
	if id != "req-123" {
		t.Errorf("GetRequestID = %q, want %q", id, "req-123")
	}
}
```

- [x] **Step 2: Run to verify it fails**

```bash
cd server && go test ./pkg/ctxbase/... -v
```

Expected: `FAIL` — package not found.

- [x] **Step 3: Implement**

`server/pkg/ctxbase/ctxbase.go`:
```go
package ctxbase

import "context"

type contextKey string

const (
	userIDKey    contextKey = "userID"
	requestIDKey contextKey = "requestID"
)

func SetUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func GetRequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd server && go test ./pkg/ctxbase/... -v
```

Expected: `PASS`.

---

## Task 4: pkg/httpbase — HTTP Response Helpers

**Files:**
- Create: `server/pkg/httpbase/httpbase.go`
- Create: `server/pkg/httpbase/httpbase_test.go`

> Note: `Success` passes `data` directly to `c.JSON` (matching existing `pkg/response` behavior — no outer `{"data": ...}` wrapper). `Error` wraps in `{"error": "..."}`.

- [x] **Step 1: Write the failing test**

`server/pkg/httpbase/httpbase_test.go`:
```go
package httpbase_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sonpham/vielish/server/pkg/httpbase"
)

func init() { gin.SetMode(gin.TestMode) }

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	httpbase.Success(c, http.StatusOK, gin.H{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["key"] != "value" {
		t.Errorf("body[key] = %q, want %q", body["key"], "value")
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	httpbase.Error(c, http.StatusBadRequest, "something went wrong")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["error"] != "something went wrong" {
		t.Errorf("body[error] = %q, want %q", body["error"], "something went wrong")
	}
}
```

- [x] **Step 2: Run to verify it fails**

```bash
cd server && go test ./pkg/httpbase/... -v
```

Expected: `FAIL` — package not found.

- [x] **Step 3: Implement**

`server/pkg/httpbase/httpbase.go`:
```go
package httpbase

import "github.com/gin-gonic/gin"

func Success(c *gin.Context, status int, data any) {
	c.JSON(status, data)
}

func Error(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}
```

- [x] **Step 4: Run tests to verify they pass**

```bash
cd server && go test ./pkg/httpbase/... -v
```

Expected: `PASS`.

---

## Task 5: pkg/log — Zap Logger

**Files:**
- Create: `server/pkg/log/log.go`

No dedicated unit test (infrastructure — depends on zap internals). Compilation verifies correctness.

- [x] **Step 1: Implement**

`server/pkg/log/log.go`:
```go
package log

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/sonpham/vielish/server/pkg/config"
)

func NewLogger(lc fx.Lifecycle, cfg config.Config) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if cfg.App.Env == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			_ = logger.Sync()
			return nil
		},
	})
	return logger, nil
}

var Module = fx.Module("log",
	fx.Provide(NewLogger),
)
```

- [x] **Step 2: Verify it compiles**

```bash
cd server && go build ./pkg/log/...
```

Expected: exits 0, no output.

---

## Task 6: pkg/tracing — OpenTelemetry

**Files:**
- Create: `server/pkg/tracing/tracing.go`

When `tracing.enabled: false` (the default), returns a no-op `TracerProvider` with no network connections.

- [x] **Step 1: Implement**

`server/pkg/tracing/tracing.go`:
```go
package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/sonpham/vielish/server/pkg/config"
)

func NewTracer(lc fx.Lifecycle, cfg config.Config, log *zap.Logger) (*sdktrace.TracerProvider, error) {
	if !cfg.Tracing.Enabled {
		tp := sdktrace.NewTracerProvider()
		otel.SetTracerProvider(tp)
		return tp, nil
	}

	ctx := context.Background()
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.Tracing.Endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
	)
	otel.SetTracerProvider(tp)
	log.Info("tracing enabled", zap.String("endpoint", cfg.Tracing.Endpoint))

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return tp.Shutdown(ctx)
		},
	})
	return tp, nil
}

var Module = fx.Module("tracing",
	fx.Provide(NewTracer),
)
```

- [x] **Step 2: Verify it compiles**

```bash
cd server && go build ./pkg/tracing/...
```

Expected: exits 0.

---

## Task 7: internal/driven/database — GORM + Redis Connections

**Files:**
- Create: `server/internal/driven/database/gorm.go`
- Create: `server/internal/driven/database/redis.go`
- Create: `server/internal/driven/database/module.go`

- [x] **Step 1: Implement gorm.go**

`server/internal/driven/database/gorm.go`:
```go
package database

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/sonpham/vielish/server/pkg/config"
)

func NewGorm(cfg config.Config, log *zap.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("opening gorm db: %w", err)
	}
	log.Info("connected to postgres")
	return db, nil
}
```

- [x] **Step 2: Implement redis.go**

`server/internal/driven/database/redis.go`:
```go
package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/sonpham/vielish/server/pkg/config"
)

func NewRedis(lc fx.Lifecycle, cfg config.Config, log *zap.Logger) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis url: %w", err)
	}
	client := redis.NewClient(opt)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("closing redis connection")
			return client.Close()
		},
	})
	log.Info("connected to redis")
	return client, nil
}
```

- [x] **Step 3: Implement module.go**

`server/internal/driven/database/module.go`:
```go
package database

import "go.uber.org/fx"

var Module = fx.Module("database",
	fx.Provide(NewGorm),
	fx.Provide(NewRedis),
)
```

- [x] **Step 4: Verify it compiles**

```bash
cd server && go build ./internal/driven/database/...
```

Expected: exits 0.

---

## Task 8: internal/domain/auth — Entity, Errors, Repository Interface, Domain Service

**Files:**
- Create: `server/internal/domain/auth/entity.go`
- Create: `server/internal/domain/auth/errors.go`
- Create: `server/internal/domain/auth/repository.go`
- Create: `server/internal/domain/auth/service.go`
- Create: `server/internal/domain/auth/service_test.go`
- Create: `server/internal/domain/auth/module.go`

- [x] **Step 1: Write the failing tests**

`server/internal/domain/auth/service_test.go`:
```go
package domain_test

import (
	"testing"
	"time"

	domain "github.com/sonpham/vielish/server/internal/domain/auth"
	"github.com/sonpham/vielish/server/pkg/config"
)

func newTestService() *domain.Service {
	return domain.NewService(config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret-key-32-bytes-minimum",
			AccessTTL: time.Hour,
		},
	})
}

func TestHashPassword_CheckPassword(t *testing.T) {
	svc := newTestService()

	hash, err := svc.HashPassword("mypassword123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if err := svc.CheckPassword(hash, "mypassword123"); err != nil {
		t.Errorf("CheckPassword correct password: error = %v", err)
	}
	if err := svc.CheckPassword(hash, "wrongpassword"); err == nil {
		t.Error("CheckPassword wrong password: expected error, got nil")
	}
}

func TestGenerateAccessToken_ValidateAccessToken(t *testing.T) {
	svc := newTestService()

	token, err := svc.GenerateAccessToken("user-xyz")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("GenerateAccessToken() returned empty token")
	}

	userID, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}
	if userID != "user-xyz" {
		t.Errorf("ValidateAccessToken() userID = %q, want %q", userID, "user-xyz")
	}
}

func TestValidateAccessToken_Invalid(t *testing.T) {
	svc := newTestService()

	_, err := svc.ValidateAccessToken("not.a.token")
	if err == nil {
		t.Error("ValidateAccessToken invalid token: expected error, got nil")
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	svc := newTestService()
	other := domain.NewService(config.Config{
		JWT: config.JWTConfig{Secret: "different-secret", AccessTTL: time.Hour},
	})

	token, _ := svc.GenerateAccessToken("user-abc")
	_, err := other.ValidateAccessToken(token)
	if err == nil {
		t.Error("ValidateAccessToken wrong secret: expected error, got nil")
	}
}

func TestGenerateRefreshToken_Unique(t *testing.T) {
	svc := newTestService()

	tok1, err := svc.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	tok2, _ := svc.GenerateRefreshToken()

	if tok1 == tok2 {
		t.Error("GenerateRefreshToken() should produce unique tokens")
	}
	if len(tok1) != 64 {
		t.Errorf("GenerateRefreshToken() len = %d, want 64 (32 hex bytes)", len(tok1))
	}
}
```

- [x] **Step 2: Run to verify it fails**

```bash
cd server && go test ./internal/domain/auth/... -v
```

Expected: `FAIL` — package not found.

- [x] **Step 3: Implement entity.go**

`server/internal/domain/auth/entity.go`:
```go
package domain

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	Level        string
	CreatedAt    time.Time
}
```

- [x] **Step 4: Implement errors.go**

`server/internal/domain/auth/errors.go`:
```go
package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)
```

- [x] **Step 5: Implement repository.go**

`server/internal/domain/auth/repository.go`:
```go
package domain

import (
	"context"
	"time"
)

// Repository is the only interface the domain owns for user persistence.
// Refresh token operations are included here for MVP simplicity.
type Repository interface {
	Create(ctx context.Context, email, passwordHash, displayName string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	StoreRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error
	GetRefreshToken(ctx context.Context, token string) (string, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}
```

- [x] **Step 6: Implement service.go**

`server/internal/domain/auth/service.go`:
```go
package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/sonpham/vielish/server/pkg/config"
)

type Service struct {
	jwtSecret []byte
	accessTTL time.Duration
}

func NewService(cfg config.Config) *Service {
	return &Service{
		jwtSecret: []byte(cfg.JWT.Secret),
		accessTTL: cfg.JWT.AccessTTL,
	}
}

func (s *Service) HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}

func (s *Service) CheckPassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}

func (s *Service) GenerateAccessToken(userID string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(s.accessTTL).Unix(),
	})
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("signing access token: %w", err)
	}
	return signed, nil
}

func (s *Service) ValidateAccessToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}
	userID, ok := claims["sub"].(string)
	if !ok {
		return "", ErrInvalidToken
	}
	return userID, nil
}

func (s *Service) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// AccessTTLSeconds returns the access token TTL in seconds (for ExpiresIn response field).
func (s *Service) AccessTTLSeconds() int {
	return int(s.accessTTL.Seconds())
}
```

- [x] **Step 7: Implement module.go**

`server/internal/domain/auth/module.go`:
```go
package domain

import "go.uber.org/fx"

var Module = fx.Module("auth-domain",
	fx.Provide(NewService),
)
```

- [x] **Step 8: Run tests to verify they pass**

```bash
cd server && go test ./internal/domain/auth/... -v
```

Expected: all 5 tests `PASS`.

---

## Task 9: internal/appcore/auth — DTOs + UseCase

**Files:**
- Create: `server/internal/appcore/auth/dto.go`
- Create: `server/internal/appcore/auth/usecase.go`
- Create: `server/internal/appcore/auth/usecase_test.go`
- Create: `server/internal/appcore/auth/module.go`

- [x] **Step 1: Write the failing tests**

`server/internal/appcore/auth/usecase_test.go`:
```go
package appcore_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	appcore "github.com/sonpham/vielish/server/internal/appcore/auth"
	domain "github.com/sonpham/vielish/server/internal/domain/auth"
	"github.com/sonpham/vielish/server/pkg/config"
)

// mockRepo is an in-memory implementation of domain.Repository.
type mockRepo struct {
	users  []*domain.User
	tokens map[string]string
}

func newMockRepo() *mockRepo {
	return &mockRepo{tokens: make(map[string]string)}
}

func (m *mockRepo) Create(_ context.Context, email, passwordHash, displayName string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return nil, domain.ErrEmailExists
		}
	}
	u := &domain.User{
		ID:           fmt.Sprintf("user-%d", len(m.users)+1),
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		Level:        "beginner",
		CreatedAt:    time.Now(),
	}
	m.users = append(m.users, u)
	return u, nil
}

func (m *mockRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockRepo) StoreRefreshToken(_ context.Context, token, userID string, _ time.Duration) error {
	m.tokens[token] = userID
	return nil
}

func (m *mockRepo) GetRefreshToken(_ context.Context, token string) (string, error) {
	userID, ok := m.tokens[token]
	if !ok {
		return "", errors.New("token not found")
	}
	return userID, nil
}

func (m *mockRepo) DeleteRefreshToken(_ context.Context, token string) error {
	delete(m.tokens, token)
	return nil
}

func newTestUseCase() (*appcore.UseCase, *mockRepo) {
	repo := newMockRepo()
	cfg := config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key-for-unit-tests",
			AccessTTL:  time.Hour,
			RefreshTTL: 7 * 24 * time.Hour,
		},
	}
	svc := domain.NewService(cfg)
	return appcore.NewUseCase(repo, svc, cfg), repo
}

func TestUseCase_Register_Success(t *testing.T) {
	uc, _ := newTestUseCase()
	out, err := uc.Register(context.Background(), appcore.RegisterInput{
		Email: "user@example.com", Password: "pass1234", DisplayName: "User",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if out.AccessToken == "" || out.RefreshToken == "" {
		t.Error("Register() returned empty tokens")
	}
	if out.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn = %d, want 3600", out.ExpiresIn)
	}
}

func TestUseCase_Register_DuplicateEmail(t *testing.T) {
	uc, _ := newTestUseCase()
	input := appcore.RegisterInput{Email: "dup@example.com", Password: "pass1234", DisplayName: "A"}
	uc.Register(context.Background(), input)
	_, err := uc.Register(context.Background(), input)
	if !errors.Is(err, domain.ErrEmailExists) {
		t.Errorf("expected ErrEmailExists, got %v", err)
	}
}

func TestUseCase_Login_Success(t *testing.T) {
	uc, _ := newTestUseCase()
	uc.Register(context.Background(), appcore.RegisterInput{
		Email: "login@example.com", Password: "pass1234", DisplayName: "User",
	})
	out, err := uc.Login(context.Background(), appcore.LoginInput{
		Email: "login@example.com", Password: "pass1234",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if out.AccessToken == "" {
		t.Error("Login() returned empty AccessToken")
	}
}

func TestUseCase_Login_WrongPassword(t *testing.T) {
	uc, _ := newTestUseCase()
	uc.Register(context.Background(), appcore.RegisterInput{
		Email: "wp@example.com", Password: "pass1234", DisplayName: "User",
	})
	_, err := uc.Login(context.Background(), appcore.LoginInput{
		Email: "wp@example.com", Password: "wrongpass",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUseCase_Refresh_Success(t *testing.T) {
	uc, _ := newTestUseCase()
	reg, _ := uc.Register(context.Background(), appcore.RegisterInput{
		Email: "ref@example.com", Password: "pass1234", DisplayName: "User",
	})
	out, err := uc.Refresh(context.Background(), reg.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if out.RefreshToken == reg.RefreshToken {
		t.Error("Refresh() should issue a new refresh token")
	}
}

func TestUseCase_Logout(t *testing.T) {
	uc, repo := newTestUseCase()
	reg, _ := uc.Register(context.Background(), appcore.RegisterInput{
		Email: "out@example.com", Password: "pass1234", DisplayName: "User",
	})
	if err := uc.Logout(context.Background(), reg.RefreshToken); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	_, err := repo.GetRefreshToken(context.Background(), reg.RefreshToken)
	if err == nil {
		t.Error("refresh token should be deleted after Logout()")
	}
}
```

- [x] **Step 2: Run to verify it fails**

```bash
cd server && go test ./internal/appcore/auth/... -v
```

Expected: `FAIL` — package not found.

- [x] **Step 3: Implement dto.go**

`server/internal/appcore/auth/dto.go`:
```go
package appcore

type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

type LoginInput struct {
	Email    string
	Password string
}

type TokenOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}
```

- [x] **Step 4: Implement usecase.go**

`server/internal/appcore/auth/usecase.go`:
```go
package appcore

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/sonpham/vielish/server/internal/domain/auth"
	"github.com/sonpham/vielish/server/pkg/config"
)

type UseCase struct {
	repo       domain.Repository
	service    *domain.Service
	refreshTTL time.Duration
}

func NewUseCase(repo domain.Repository, service *domain.Service, cfg config.Config) *UseCase {
	return &UseCase{
		repo:       repo,
		service:    service,
		refreshTTL: cfg.JWT.RefreshTTL,
	}
}

func (uc *UseCase) Register(ctx context.Context, input RegisterInput) (*TokenOutput, error) {
	hash, err := uc.service.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}
	user, err := uc.repo.Create(ctx, input.Email, hash, input.DisplayName)
	if err != nil {
		return nil, err
	}
	return uc.issueTokens(ctx, user.ID)
}

func (uc *UseCase) Login(ctx context.Context, input LoginInput) (*TokenOutput, error) {
	user, err := uc.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if err := uc.service.CheckPassword(user.PasswordHash, input.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	return uc.issueTokens(ctx, user.ID)
}

func (uc *UseCase) Refresh(ctx context.Context, refreshToken string) (*TokenOutput, error) {
	userID, err := uc.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	if err := uc.repo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("deleting old refresh token: %w", err)
	}
	return uc.issueTokens(ctx, userID)
}

func (uc *UseCase) Logout(ctx context.Context, refreshToken string) error {
	return uc.repo.DeleteRefreshToken(ctx, refreshToken)
}

func (uc *UseCase) issueTokens(ctx context.Context, userID string) (*TokenOutput, error) {
	accessToken, err := uc.service.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}
	refreshToken, err := uc.service.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	if err := uc.repo.StoreRefreshToken(ctx, refreshToken, userID, uc.refreshTTL); err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}
	return &TokenOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    uc.service.AccessTTLSeconds(),
	}, nil
}
```

- [x] **Step 5: Implement module.go**

`server/internal/appcore/auth/module.go`:
```go
package appcore

import "go.uber.org/fx"

var Module = fx.Module("auth-appcore",
	fx.Provide(NewUseCase),
)
```

- [x] **Step 6: Run tests to verify they pass**

```bash
cd server && go test ./internal/appcore/auth/... -v
```

Expected: all 6 tests `PASS`.

---

## Task 10: internal/driven/auth — GORM Model + Repository Implementation

**Files:**
- Create: `server/internal/driven/auth/gorm_model.go`
- Create: `server/internal/driven/auth/repository.go`
- Create: `server/internal/driven/auth/module.go`

> Integration tests for this layer require a running Postgres + Redis. They are not in scope for unit testing — verify by running the full server in Task 14.

- [x] **Step 1: Implement gorm_model.go**

`server/internal/driven/auth/gorm_model.go`:
```go
package driven

import (
	"time"

	domain "github.com/sonpham/vielish/server/internal/domain/auth"
)

type UserModel struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	DisplayName  string    `gorm:"not null"`
	Level        string    `gorm:"default:'beginner'"`
	CreatedAt    time.Time
}

func (UserModel) TableName() string { return "users" }

func (m *UserModel) ToEntity() *domain.User {
	return &domain.User{
		ID:           m.ID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		DisplayName:  m.DisplayName,
		Level:        m.Level,
		CreatedAt:    m.CreatedAt,
	}
}
```

- [x] **Step 2: Implement repository.go**

`server/internal/driven/auth/repository.go`:
```go
package driven

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	domain "github.com/sonpham/vielish/server/internal/domain/auth"
)

const refreshTokenPrefix = "refresh:"

type Repository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewRepository(db *gorm.DB, redis *redis.Client) *Repository {
	return &Repository{db: db, redis: redis}
}

func (r *Repository) Create(ctx context.Context, email, passwordHash, displayName string) (*domain.User, error) {
	m := &UserModel{Email: email, PasswordHash: passwordHash, DisplayName: displayName}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, domain.ErrEmailExists
		}
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return m.ToEntity(), nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by email: %w", err)
	}
	return m.ToEntity(), nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by id: %w", err)
	}
	return m.ToEntity(), nil
}

func (r *Repository) StoreRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	return r.redis.Set(ctx, refreshTokenPrefix+token, userID, ttl).Err()
}

func (r *Repository) GetRefreshToken(ctx context.Context, token string) (string, error) {
	userID, err := r.redis.Get(ctx, refreshTokenPrefix+token).Result()
	if errors.Is(err, redis.Nil) {
		return "", domain.ErrInvalidToken
	}
	if err != nil {
		return "", fmt.Errorf("getting refresh token: %w", err)
	}
	return userID, nil
}

func (r *Repository) DeleteRefreshToken(ctx context.Context, token string) error {
	return r.redis.Del(ctx, refreshTokenPrefix+token).Err()
}
```

- [x] **Step 3: Implement module.go**

`server/internal/driven/auth/module.go`:
```go
package driven

import (
	"go.uber.org/fx"

	domain "github.com/sonpham/vielish/server/internal/domain/auth"
)

var Module = fx.Module("auth-driven",
	fx.Provide(
		fx.Annotate(
			NewRepository,
			fx.As(new(domain.Repository)),
		),
	),
)
```

- [x] **Step 4: Verify it compiles**

```bash
cd server && go build ./internal/driven/auth/...
```

Expected: exits 0.

---

## Task 11: internal/driving/httpui — Handler, Presenter, Middleware, Server

**Files:**
- Create: `server/internal/driving/httpui/middleware/auth.go`
- Create: `server/internal/driving/httpui/middleware/auth_test.go`
- Create: `server/internal/driving/httpui/presenter/auth_presenter.go`
- Create: `server/internal/driving/httpui/handler/auth_handler.go`
- Create: `server/internal/driving/httpui/handler/auth_handler_test.go`
- Create: `server/internal/driving/httpui/server.go`
- Create: `server/internal/driving/httpui/module.go`

- [x] **Step 1: Write failing tests for middleware**

`server/internal/driving/httpui/middleware/auth_test.go`:
```go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	appcore "github.com/sonpham/vielish/server/internal/appcore/auth"
	domain "github.com/sonpham/vielish/server/internal/domain/auth"
	"github.com/sonpham/vielish/server/internal/driving/httpui/middleware"
	"github.com/sonpham/vielish/server/pkg/config"
	"github.com/sonpham/vielish/server/pkg/ctxbase"
)

func init() { gin.SetMode(gin.TestMode) }

func newTestService() *domain.Service {
	return domain.NewService(config.Config{
		JWT: config.JWTConfig{Secret: "test-secret", AccessTTL: time.Hour},
	})
}

func TestAuthMiddleware_Valid(t *testing.T) {
	svc := newTestService()
	token, _ := svc.GenerateAccessToken("user-999")

	r := gin.New()
	r.Use(middleware.Auth(svc))
	r.GET("/protected", func(c *gin.Context) {
		id, ok := ctxbase.GetUserID(c.Request.Context())
		if !ok || id != "user-999" {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	svc := newTestService()
	r := gin.New()
	r.Use(middleware.Auth(svc))
	r.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	svc := newTestService()
	r := gin.New()
	r.Use(middleware.Auth(svc))
	r.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer bad.token.here")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// Ensure middleware_test compiles even without handler package dependency issues
var _ = appcore.TokenOutput{}
```

- [x] **Step 2: Run to verify middleware test fails**

```bash
cd server && go test ./internal/driving/httpui/middleware/... -v
```

Expected: `FAIL` — package not found.

- [x] **Step 3: Implement middleware/auth.go**

`server/internal/driving/httpui/middleware/auth.go`:
```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	domain "github.com/sonpham/vielish/server/internal/domain/auth"
	"github.com/sonpham/vielish/server/pkg/ctxbase"
)

func Auth(svc *domain.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}
		userID, err := svc.ValidateAccessToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
		ctx := ctxbase.SetUserID(c.Request.Context(), userID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
```

- [x] **Step 4: Run middleware tests to verify they pass**

```bash
cd server && go test ./internal/driving/httpui/middleware/... -v
```

Expected: all 3 middleware tests `PASS`.

- [x] **Step 5: Write failing handler tests**

`server/internal/driving/httpui/handler/auth_handler_test.go`:
```go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	appcore "github.com/sonpham/vielish/server/internal/appcore/auth"
	domain "github.com/sonpham/vielish/server/internal/domain/auth"
	"github.com/sonpham/vielish/server/internal/driving/httpui/handler"
	"github.com/sonpham/vielish/server/internal/driving/httpui/presenter"
)

func init() { gin.SetMode(gin.TestMode) }

// stubUseCase lets tests control UseCase behaviour.
type stubUseCase struct {
	registerFn func(ctx context.Context, in appcore.RegisterInput) (*appcore.TokenOutput, error)
	loginFn    func(ctx context.Context, in appcore.LoginInput) (*appcore.TokenOutput, error)
	refreshFn  func(ctx context.Context, token string) (*appcore.TokenOutput, error)
	logoutFn   func(ctx context.Context, token string) error
}

func (s *stubUseCase) Register(ctx context.Context, in appcore.RegisterInput) (*appcore.TokenOutput, error) {
	return s.registerFn(ctx, in)
}
func (s *stubUseCase) Login(ctx context.Context, in appcore.LoginInput) (*appcore.TokenOutput, error) {
	return s.loginFn(ctx, in)
}
func (s *stubUseCase) Refresh(ctx context.Context, token string) (*appcore.TokenOutput, error) {
	return s.refreshFn(ctx, token)
}
func (s *stubUseCase) Logout(ctx context.Context, token string) error {
	return s.logoutFn(ctx, token)
}
```

> **Note:** The handler currently accepts `*appcore.UseCase` (concrete type). To enable the stub above, `handler.NewHandler` must accept an interface. Update the handler to use `UseCaseInterface` (see Step 6 below).

- [x] **Step 6: Implement handler with use-case interface**

`server/internal/driving/httpui/handler/auth_handler.go`:
```go
package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	appcore "github.com/sonpham/vielish/server/internal/appcore/auth"
	domain "github.com/sonpham/vielish/server/internal/domain/auth"
	"github.com/sonpham/vielish/server/internal/driving/httpui/presenter"
	"github.com/sonpham/vielish/server/pkg/httpbase"
)

// UseCaseInterface allows the handler to be tested with stubs.
type UseCaseInterface interface {
	Register(ctx context.Context, in appcore.RegisterInput) (*appcore.TokenOutput, error)
	Login(ctx context.Context, in appcore.LoginInput) (*appcore.TokenOutput, error)
	Refresh(ctx context.Context, token string) (*appcore.TokenOutput, error)
	Logout(ctx context.Context, token string) error
}

type registerRequest struct {
	Email       string `json:"email"        binding:"required,email"`
	Password    string `json:"password"     binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type Handler struct {
	useCase   UseCaseInterface
	presenter *presenter.AuthPresenter
}

func NewHandler(uc *appcore.UseCase, p *presenter.AuthPresenter) *Handler {
	return &Handler{useCase: uc, presenter: p}
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpbase.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}
	out, err := h.useCase.Register(c.Request.Context(), appcore.RegisterInput{
		Email: req.Email, Password: req.Password, DisplayName: req.DisplayName,
	})
	if errors.Is(err, domain.ErrEmailExists) {
		httpbase.Error(c, http.StatusConflict, "email already registered")
		return
	}
	if err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "registration failed")
		return
	}
	h.presenter.Tokens(c, http.StatusCreated, out)
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpbase.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}
	out, err := h.useCase.Login(c.Request.Context(), appcore.LoginInput{
		Email: req.Email, Password: req.Password,
	})
	if errors.Is(err, domain.ErrInvalidCredentials) {
		httpbase.Error(c, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "login failed")
		return
	}
	h.presenter.Tokens(c, http.StatusOK, out)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpbase.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}
	out, err := h.useCase.Refresh(c.Request.Context(), req.RefreshToken)
	if errors.Is(err, domain.ErrInvalidToken) {
		httpbase.Error(c, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	if err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "refresh failed")
		return
	}
	h.presenter.Tokens(c, http.StatusOK, out)
}

func (h *Handler) Logout(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpbase.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}
	if err := h.useCase.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "logout failed")
		return
	}
	httpbase.Success(c, http.StatusOK, gin.H{"message": "logged out"})
}
```

- [x] **Step 7: Implement presenter/auth_presenter.go**

`server/internal/driving/httpui/presenter/auth_presenter.go`:
```go
package presenter

import (
	"github.com/gin-gonic/gin"

	appcore "github.com/sonpham/vielish/server/internal/appcore/auth"
	"github.com/sonpham/vielish/server/pkg/httpbase"
)

type AuthPresenter struct{}

func NewAuthPresenter() *AuthPresenter { return &AuthPresenter{} }

func (p *AuthPresenter) Tokens(c *gin.Context, status int, out *appcore.TokenOutput) {
	httpbase.Success(c, status, gin.H{
		"access_token":  out.AccessToken,
		"refresh_token": out.RefreshToken,
		"expires_in":    out.ExpiresIn,
	})
}
```

- [x] **Step 8: Complete the handler test**

Add these test functions to `auth_handler_test.go`:
```go
func newTestHandler(stub *stubUseCase) (*handler.Handler, *gin.Engine) {
	p := presenter.NewAuthPresenter()
	h := handler.NewHandlerFromInterface(stub, p)
	r := gin.New()
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	r.POST("/refresh", h.Refresh)
	r.POST("/logout", h.Logout)
	return h, r
}

func postJSON(t *testing.T, r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestHandler_Register_201(t *testing.T) {
	stub := &stubUseCase{
		registerFn: func(_ context.Context, in appcore.RegisterInput) (*appcore.TokenOutput, error) {
			return &appcore.TokenOutput{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600}, nil
		},
	}
	_, r := newTestHandler(stub)
	w := postJSON(t, r, "/register", map[string]string{
		"email": "a@b.com", "password": "pass1234", "display_name": "A",
	})
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201; body: %s", w.Code, w.Body)
	}
}

func TestHandler_Register_409_DuplicateEmail(t *testing.T) {
	stub := &stubUseCase{
		registerFn: func(_ context.Context, _ appcore.RegisterInput) (*appcore.TokenOutput, error) {
			return nil, domain.ErrEmailExists
		},
	}
	_, r := newTestHandler(stub)
	w := postJSON(t, r, "/register", map[string]string{
		"email": "dup@b.com", "password": "pass1234", "display_name": "A",
	})
	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", w.Code)
	}
}

func TestHandler_Login_401_BadPassword(t *testing.T) {
	stub := &stubUseCase{
		loginFn: func(_ context.Context, _ appcore.LoginInput) (*appcore.TokenOutput, error) {
			return nil, domain.ErrInvalidCredentials
		},
	}
	_, r := newTestHandler(stub)
	w := postJSON(t, r, "/login", map[string]string{"email": "a@b.com", "password": "wrong"})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}
```

Also add `NewHandlerFromInterface` to `auth_handler.go` to make the stub work:
```go
// NewHandlerFromInterface is used in tests to inject a stubbed use case.
func NewHandlerFromInterface(uc UseCaseInterface, p *presenter.AuthPresenter) *Handler {
	return &Handler{useCase: uc, presenter: p}
}
```

- [x] **Step 9: Run handler tests**

```bash
cd server && go test ./internal/driving/httpui/handler/... -v
```

Expected: `TestHandler_Register_201`, `TestHandler_Register_409_DuplicateEmail`, `TestHandler_Login_401_BadPassword` all `PASS`.

- [x] **Step 10: Implement server.go**

`server/internal/driving/httpui/server.go`:
```go
package httpui

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	domain "github.com/sonpham/vielish/server/internal/domain/auth"
	"github.com/sonpham/vielish/server/internal/driving/httpui/handler"
	"github.com/sonpham/vielish/server/internal/driving/httpui/middleware"
	"github.com/sonpham/vielish/server/pkg/config"
)

func NewGin(cfg config.Config) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	return gin.New()
}

func RegisterRoutes(r *gin.Engine, h *handler.Handler, svc *domain.Service, cfg config.Config) {
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.Origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	auth := r.Group("/api/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)
	}

	// Protected group — placeholder for vocab, listening features.
	_ = r.Group("/api").Use(middleware.Auth(svc))
}

func RegisterLifecycle(lc fx.Lifecycle, r *gin.Engine, cfg config.Config, log *zap.Logger) {
	srv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: r,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("starting HTTP server", zap.String("port", cfg.App.Port))
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error("HTTP server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("shutting down HTTP server")
			return srv.Shutdown(ctx)
		},
	})
}
```

- [x] **Step 11: Implement module.go**

`server/internal/driving/httpui/module.go`:
```go
package httpui

import (
	"go.uber.org/fx"

	"github.com/sonpham/vielish/server/internal/driving/httpui/handler"
	"github.com/sonpham/vielish/server/internal/driving/httpui/presenter"
)

var Module = fx.Module("httpui",
	fx.Provide(NewGin),
	fx.Provide(handler.NewHandler),
	fx.Provide(presenter.NewAuthPresenter),
	fx.Invoke(RegisterRoutes),
	fx.Invoke(RegisterLifecycle),
)
```

- [x] **Step 12: Verify all httpui compiles**

```bash
cd server && go build ./internal/driving/...
```

Expected: exits 0.

---

## Task 12: cmd/api/main.go + config.yaml — fx Wiring

**Files:**
- Create: `server/cmd/api/config.yaml`
- Rewrite: `server/cmd/api/main.go`

- [x] **Step 1: Create config.yaml**

`server/cmd/api/config.yaml`:
```yaml
app:
  port: "8080"
  env: development

database:
  url: postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable

redis:
  url: redis://localhost:6379

jwt:
  secret: dev-secret-change-in-production
  access_ttl: 1h
  refresh_ttl: 168h

cors:
  origins:
    - http://localhost:3000

tracing:
  enabled: false
  endpoint: localhost:4317
```

- [x] **Step 2: Rewrite main.go**

`server/cmd/api/main.go`:
```go
package main

import (
	"go.uber.org/fx"

	authAppcore "github.com/sonpham/vielish/server/internal/appcore/auth"
	"github.com/sonpham/vielish/server/internal/driven/database"
	authDriven "github.com/sonpham/vielish/server/internal/driven/auth"
	authDomain "github.com/sonpham/vielish/server/internal/domain/auth"
	"github.com/sonpham/vielish/server/internal/driving/httpui"
	"github.com/sonpham/vielish/server/pkg/config"
	pkglog "github.com/sonpham/vielish/server/pkg/log"
	"github.com/sonpham/vielish/server/pkg/tracing"
)

func main() {
	fx.New(
		// Infrastructure
		config.Module,
		pkglog.Module,
		tracing.Module,
		database.Module,

		// Auth feature
		fx.Module("auth",
			authDomain.Module,
			authAppcore.Module,
			authDriven.Module,
		),

		// HTTP server
		httpui.Module,
	).Run()
}
```

- [x] **Step 3: Verify the binary compiles (old packages still exist — that's fine for now)**

```bash
cd server && go build ./cmd/api/...
```

Expected: exits 0. If there are import conflicts with the old packages, proceed to Task 13 before re-trying.

---

## Task 13: Delete Old Code

Remove all packages superseded by the new architecture.

- [x] **Step 1: Delete old packages**

```bash
rm -rf server/internal/auth/
rm -rf server/internal/config/
rm -rf server/internal/database/
rm -rf server/internal/router/
rm -rf server/pkg/response/
```

- [x] **Step 2: Tidy modules**

```bash
cd server && go mod tidy
```

Expected: exits 0. No references to deleted packages remain.

- [x] **Step 3: Build and test**

```bash
cd server && go build ./... && go test ./...
```

Expected: build exits 0; all unit tests pass. DB-dependent tests are skipped (no `TEST_DATABASE_URL` set in CI by default).

---

## Task 14: Build and Smoke Test

Requires running Postgres + Redis (use existing docker-compose).

- [x] **Step 1: Start infrastructure**

```bash
cd /path/to/vielish && docker-compose up -d postgres redis
```

Expected: both containers healthy.

- [x] **Step 2: Run the server**

```bash
cd server && go run cmd/api/main.go
```

Expected output (zap development logger):
```
{"level":"info","msg":"connected to postgres"}
{"level":"info","msg":"connected to redis"}
{"level":"info","msg":"starting HTTP server","port":"8080"}
```

- [x] **Step 3: Health check**

```bash
curl http://localhost:8080/api/health
```

Expected: `{"status":"ok"}`

- [x] **Step 4: Register user**

```bash
curl -s -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"son@example.com","password":"password123","display_name":"Son"}' | jq .
```

Expected (HTTP 201):
```json
{
  "access_token": "<jwt>",
  "refresh_token": "<hex>",
  "expires_in": 3600
}
```

- [x] **Step 5: Login**

```bash
curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"son@example.com","password":"password123"}' | jq .
```

Expected: HTTP 200 with tokens.

- [x] **Step 6: Duplicate email → 409**

```bash
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"son@example.com","password":"password123","display_name":"Son2"}'
```

Expected: `409`

- [x] **Step 7: Wrong password → 401**

```bash
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"son@example.com","password":"wrongpassword"}'
```

Expected: `401`

---

## Adding Future Features (vocab, listening)

Adding a new domain (e.g., vocab) follows the same pattern. Create four packages:
1. `internal/domain/vocab/` — entity, repository interface, domain service, errors, module
2. `internal/appcore/vocab/` — DTOs, use case, module
3. `internal/driven/vocab/` — GORM model, repository impl, module
4. `internal/driving/httpui/handler/vocab_handler.go` + routes in `server.go`

Then add to `main.go`:
```go
fx.Module("vocab",
    vocabDomain.Module,
    vocabAppcore.Module,
    vocabDriven.Module,
),
```

No changes to existing modules required.
