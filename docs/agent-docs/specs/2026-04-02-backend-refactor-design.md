# Backend Refactor Design — DDD/Clean Architecture

**Date:** 2026-04-02
**Scope:** Refactor `server/` to follow DDD/Clean Architecture with Uber fx, GORM, Viper, Zap, and OpenTelemetry.

---

## Overview

Refactor the Go backend from a flat per-feature package layout to a layered DDD/Clean Architecture. The refactor is big-bang (all at once). The existing auth module serves as the reference implementation for all future features (vocab, listening).

---

## Folder Structure

```
server/
├── cmd/api/
│   ├── main.go              # fx.New(...) entrypoint
│   └── config.yaml          # Viper config file
├── internal/
│   ├── appcore/             # Application layer: UseCases + DTOs
│   │   └── auth/
│   │       ├── usecase.go
│   │       ├── dto.go
│   │       └── module.go    # fx.Module("auth-appcore", ...)
│   ├── domain/              # Domain layer: Entities + Interfaces + Domain Services
│   │   └── auth/
│   │       ├── entity.go
│   │       ├── repository.go
│   │       ├── service.go
│   │       ├── errors.go
│   │       └── module.go
│   ├── driven/              # Driven adapters: DB, Redis
│   │   ├── database/        # Shared DB/Redis connection setup
│   │   │   ├── gorm.go      # NewGorm(*gorm.DB) from config
│   │   │   ├── redis.go     # NewRedis(*redis.Client) from config
│   │   │   └── module.go
│   │   └── auth/
│   │       ├── gorm_model.go
│   │       ├── repository.go
│   │       └── module.go
│   └── driving/
│       └── httpui/
│           ├── handler/
│           │   └── auth_handler.go
│           ├── presenter/
│           │   └── auth_presenter.go
│           ├── middleware/
│           │   └── auth.go
│           └── server.go
├── pkg/
│   ├── config/              # Viper config loader
│   ├── ctxbase/             # Context keys and helpers
│   ├── httpbase/            # HTTP response helpers
│   ├── log/                 # Zap logger
│   └── tracing/             # OpenTelemetry
└── migrations/              # Unchanged
```

---

## Layer Responsibilities

| Layer | Package | Responsibility |
|---|---|---|
| Domain | `internal/domain/auth/` | User entity, repository interface, JWT/bcrypt domain service, domain errors |
| Application | `internal/appcore/auth/` | AuthUseCase orchestrates domain service + repo, DTOs |
| Driven | `internal/driven/auth/` | GORM UserModel, repository implementation (DB + Redis) |
| Driving | `internal/driving/httpui/` | Gin handlers, presenters, middleware, router |
| Shared | `pkg/` | Config, logging, tracing, HTTP helpers, context utilities |

---

## Domain Layer — `internal/domain/auth/`

### `entity.go`
Pure Go struct, no framework dependencies, no GORM tags.

```go
type User struct {
    ID           string
    Email        string
    PasswordHash string
    DisplayName  string
    Level        string
    CreatedAt    time.Time
}
```

### `service.go`
Stateless domain service. No DB or Redis calls.

```go
type Service struct { jwtSecret []byte }

func NewService(cfg config.Config) *Service
func (s *Service) HashPassword(plain string) (string, error)
func (s *Service) CheckPassword(hash, plain string) error
func (s *Service) GenerateAccessToken(userID string) (string, error)
func (s *Service) ValidateAccessToken(token string) (string, error) // returns userID
func (s *Service) GenerateRefreshToken() (string, error)            // random hex
```

### `repository.go`
Interface only. Domain has no knowledge of GORM or Redis.

```go
type Repository interface {
    Create(ctx context.Context, email, passwordHash, displayName string) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    GetByID(ctx context.Context, id string) (*User, error)
    StoreRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error
    GetRefreshToken(ctx context.Context, token string) (string, error)
    DeleteRefreshToken(ctx context.Context, token string) error
}
```

> Refresh token operations are included in Repository for MVP simplicity. Can be split into a separate `SessionRepository` when needed.

### `errors.go`
```go
var (
    ErrUserNotFound       = errors.New("user not found")
    ErrEmailExists        = errors.New("email already exists")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrInvalidToken       = errors.New("invalid token")
)
```

---

## Application Layer — `internal/appcore/auth/`

### `dto.go`
Request/response data objects for the application boundary.

```go
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

### `usecase.go`
Orchestrates domain service and repository. Contains no business logic itself.

```go
type UseCase struct {
    repo    domain.Repository
    service *domain.Service
}

func NewUseCase(repo domain.Repository, service *domain.Service) *UseCase
func (uc *UseCase) Register(ctx context.Context, input RegisterInput) (*TokenOutput, error)
func (uc *UseCase) Login(ctx context.Context, input LoginInput) (*TokenOutput, error)
func (uc *UseCase) Refresh(ctx context.Context, refreshToken string) (*TokenOutput, error)
func (uc *UseCase) Logout(ctx context.Context, refreshToken string) error
```

**Register flow:**
1. `domain.Service.HashPassword(input.Password)`
2. `domain.Repository.Create(email, hash, displayName)`
3. `domain.Service.GenerateAccessToken(user.ID)`
4. `domain.Service.GenerateRefreshToken()`
5. `domain.Repository.StoreRefreshToken(token, userID, ttl)`
6. Return `TokenOutput`

---

## Driven Layer — `internal/driven/auth/`

### `gorm_model.go`
GORM-tagged struct, completely separate from the domain entity.

```go
type UserModel struct {
    ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Email        string    `gorm:"uniqueIndex;not null"`
    PasswordHash string    `gorm:"not null"`
    DisplayName  string    `gorm:"not null"`
    Level        string    `gorm:"default:'beginner'"`
    CreatedAt    time.Time
}

func (m *UserModel) ToEntity() *domain.User
func FromEntity(u *domain.User) *UserModel
```

### `repository.go`
Implements `domain.Repository`. Receives `*gorm.DB` and `*redis.Client` via fx injection.

- DB operations use GORM
- Refresh token operations use Redis directly

---

## Driving Layer — `internal/driving/httpui/`

### `handler/auth_handler.go`
Receives `*appcore.UseCase` via fx. Binds HTTP request → calls UseCase → calls Presenter.

```go
func (h *Handler) Register(c *gin.Context)
func (h *Handler) Login(c *gin.Context)
func (h *Handler) Refresh(c *gin.Context)
func (h *Handler) Logout(c *gin.Context)
```

### `presenter/auth_presenter.go`
Maps `appcore.TokenOutput` → HTTP JSON response. Owns the HTTP response shape.

### `middleware/auth.go`
Extracts Bearer token → calls `domain.Service.ValidateAccessToken()` → sets userID in context via `ctxbase`.

### `server.go`
Gin router setup + fx lifecycle hooks for graceful shutdown.

```go
func Register(lc fx.Lifecycle, r *gin.Engine, cfg config.Config) {
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            go r.Run(":" + cfg.App.Port)
            return nil
        },
        OnStop: func(ctx context.Context) error {
            // graceful shutdown
            return nil
        },
    })
}
```

---

## `pkg/` Packages

### `pkg/config/`
- Uses Viper to load `config.yaml`
- Env vars override config file (e.g., `DATABASE_URL` overrides `database.url`)
- Single `Config` struct injected app-wide via fx

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
  origins: ["http://localhost:3000"]
tracing:
  enabled: false
  endpoint: localhost:4317
```

### `pkg/log/`
- Uses `go.uber.org/zap`
- `NewLogger(cfg Config) (*zap.Logger, error)`
- Development: console encoder with color
- Production: JSON encoder

### `pkg/tracing/`
- Uses OpenTelemetry SDK + OTLP exporter
- `NewTracer(cfg Config, log *zap.Logger) (*sdktrace.TracerProvider, error)`
- fx lifecycle: OnStart initializes exporter, OnStop flushes and shuts down

### `pkg/ctxbase/`
Typed context keys to avoid key collisions across packages.

```go
func SetUserID(ctx context.Context, id string) context.Context
func GetUserID(ctx context.Context) (string, bool)
func SetRequestID(ctx context.Context, id string) context.Context
func GetRequestID(ctx context.Context) (string, bool)
```

### `pkg/httpbase/`
HTTP response helpers (migrated from `pkg/response/`).

```go
func Success(c *gin.Context, status int, data any)
func Error(c *gin.Context, status int, msg string)
```

---

## fx Wiring — `cmd/api/main.go`

```go
fx.New(
    // Infrastructure
    config.Module,
    log.Module,
    tracing.Module,
    database.Module,  // internal/driven/database — fx.Provide(NewGorm, NewRedis)

    // Auth feature module (Option A: module per domain)
    fx.Module("auth",
        authDomain.Module,   // fx.Provide(domain.NewService)
        authAppcore.Module,  // fx.Provide(usecase.NewUseCase)
        authDriven.Module,   // fx.Provide(driven.NewRepository)
        authDriving.Module,  // fx.Provide(handler.NewHandler)
    ),

    // HTTP Server
    httpui.Module,
)
```

Each `module.go` file uses `fx.Module` to scope its providers. Adding a new feature (e.g., vocab) means adding a new `fx.Module("vocab", ...)` block.

---

## Key Decisions

- **Domain vs Data model:** `domain.User` has no GORM tags. `driven.UserModel` has GORM tags and provides `ToEntity()`/`FromEntity()` mappers.
- **Repository interface location:** In `domain/` (not `appcore/`). Domain owns its own contracts.
- **Refresh tokens:** Managed via `domain.Repository` for MVP simplicity.
- **No WebSocket, no AI features:** MVP scope only.
- **Presenter responsibility:** Owns HTTP response shape. Handler does not build JSON directly.
