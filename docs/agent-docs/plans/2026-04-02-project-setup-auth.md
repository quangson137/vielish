# Project Setup + Auth Implementation Plan

> Steps use checkbox (`- [ ]`) syntax for tracking progress.

**Goal:** Set up the full project infrastructure (Docker, PostgreSQL, Redis, Go backend, Next.js frontend) and implement JWT authentication.

**Architecture:** Go backend using Gin framework with REST API, JWT access tokens stored client-side, refresh tokens in Redis. PostgreSQL for persistent data, Redis for sessions/cache. Next.js App Router frontend with server-side and client-side auth handling.

**Tech Stack:** Go 1.22+, Gin, PostgreSQL 16, Redis 7, Next.js 14 (App Router), Docker Compose, golang-jwt, bcrypt

---

## File Structure

```
vielish/
  docker-compose.yml                    — PostgreSQL + Redis services
  .env.example                          — Environment variables template
  server/
    go.mod                              — Go module definition
    go.sum
    cmd/api/main.go                     — Server entrypoint
    internal/
      config/config.go                  — Environment configuration
      database/postgres.go              — PostgreSQL connection
      database/redis.go                 — Redis connection
      auth/handler.go                   — Auth HTTP handlers
      auth/handler_test.go              — Auth handler tests
      auth/service.go                   — Auth business logic
      auth/service_test.go              — Auth service tests
      auth/model.go                     — User model + DTOs
      auth/repository.go               — User DB queries
      auth/repository_test.go           — Repository tests
      auth/middleware.go                — JWT auth middleware
      auth/middleware_test.go           — Middleware tests
      router/router.go                  — Route registration
    migrations/
      001_create_users.up.sql           — Users table
      001_create_users.down.sql         — Drop users table
    pkg/
      response/response.go             — Standard JSON response helpers
  web/
    package.json
    next.config.js
    app/
      layout.tsx                        — Root layout
      page.tsx                          — Home/landing page
      login/page.tsx                    — Login page
      register/page.tsx                 — Register page
      dashboard/page.tsx                — Protected dashboard (placeholder)
      dashboard/layout.tsx              — Dashboard layout with auth guard
    components/
      auth-form.tsx                     — Shared login/register form
    lib/
      api.ts                            — API client with token handling
      auth-context.tsx                  — Auth context provider
```

---

### Task 1: Docker Compose Setup

**Files:**
- Create: `docker-compose.yml`

- [ ] **Step 1: Create Docker Compose file**

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: vielish
      POSTGRES_PASSWORD: vielish_dev
      POSTGRES_DB: vielish
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redisdata:/data

volumes:
  pgdata:
  redisdata:
```

- [ ] **Step 2: Start services and verify**

Run: `docker-compose up -d`
Expected: Both containers running

Run: `docker-compose ps`
Expected: postgres and redis both "Up"

Run: `docker exec -it $(docker-compose ps -q postgres) psql -U vielish -c "SELECT 1;"`
Expected: Returns `1`

---

### Task 2: Go Module + Config

**Files:**
- Create: `server/go.mod`
- Create: `server/internal/config/config.go`
- Create: `server/internal/config/config_test.go`

- [ ] **Step 1: Initialize Go module**

Run: `cd server && go mod init github.com/sonpham/vielish/server`

- [ ] **Step 2: Write config test**

```go
// server/internal/config/config_test.go
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
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd server && go test ./internal/config/ -v`
Expected: FAIL — `Load` undefined

- [ ] **Step 4: Write config implementation**

```go
// server/internal/config/config.go
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
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd server && go test ./internal/config/ -v`
Expected: PASS

---

### Task 3: Database Connections

**Files:**
- Create: `server/internal/database/postgres.go`
- Create: `server/internal/database/redis.go`

- [ ] **Step 1: Install dependencies**

Run: `cd server && go get github.com/jackc/pgx/v5/pgxpool github.com/redis/go-redis/v9`

- [ ] **Step 2: Write PostgreSQL connection**

```go
// server/internal/database/postgres.go
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}

	return pool, nil
}
```

- [ ] **Step 3: Write Redis connection**

```go
// server/internal/database/redis.go
package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedis(ctx context.Context, redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	return client, nil
}
```

- [ ] **Step 4: Verify compilation**

Run: `cd server && go build ./internal/database/`
Expected: No errors

---

### Task 4: Database Migration — Users Table

**Files:**
- Create: `server/migrations/001_create_users.up.sql`
- Create: `server/migrations/001_create_users.down.sql`

- [ ] **Step 1: Install golang-migrate CLI**

Run (macOS): `brew install golang-migrate`

Alternatively (any OS): `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

- [ ] **Step 2: Create up migration**

```sql
-- server/migrations/001_create_users.up.sql
CREATE TYPE user_level AS ENUM ('beginner', 'intermediate', 'advanced');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    level user_level NOT NULL DEFAULT 'beginner',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

- [ ] **Step 2: Create down migration**

```sql
-- server/migrations/001_create_users.down.sql
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_level;
```

- [ ] **Step 4: Apply migration using golang-migrate**

Run: `migrate -path server/migrations -database "postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable" up`
Expected: `1/u create_users (Xms)`

Run: `docker exec -i $(docker-compose ps -q postgres) psql -U vielish vielish -c "\d users"`
Expected: Table structure with id, email, password_hash, display_name, level, created_at

- [ ] **Step 5: Roll back and re-apply to verify down migration**

Run: `migrate -path server/migrations -database "postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable" down 1`
Expected: `1/d create_users (Xms)`

Run: `migrate -path server/migrations -database "postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable" up`
Expected: `1/u create_users (Xms)`

---

### Task 5: Response Helpers

**Files:**
- Create: `server/pkg/response/response.go`
- Create: `server/pkg/response/response_test.go`

- [ ] **Step 1: Write test**

```go
// server/pkg/response/response_test.go
package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, http.StatusOK, gin.H{"name": "test"})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)

	if body["name"] != "test" {
		t.Errorf("body = %v, want name=test", body)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Error(c, http.StatusBadRequest, "invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)

	if body["error"] != "invalid input" {
		t.Errorf("error = %q, want %q", body["error"], "invalid input")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd server && go get github.com/gin-gonic/gin && go test ./pkg/response/ -v`
Expected: FAIL — `Success` and `Error` undefined

- [ ] **Step 3: Write implementation**

```go
// server/pkg/response/response.go
package response

import "github.com/gin-gonic/gin"

func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd server && go test ./pkg/response/ -v`
Expected: PASS

---

### Task 6: Auth Model

**Files:**
- Create: `server/internal/auth/model.go`

- [ ] **Step 1: Write auth model and DTOs**

```go
// server/internal/auth/model.go
package auth

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"display_name"`
	Level        string    `json:"level"`
	CreatedAt    time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd server && go build ./internal/auth/`
Expected: No errors

---

### Task 7: Auth Repository

**Files:**
- Create: `server/internal/auth/repository.go`
- Create: `server/internal/auth/repository_test.go`

- [ ] **Step 1: Write repository test**

Note: These tests require a running PostgreSQL instance (integration tests).

```go
// server/internal/auth/repository_test.go
package auth

import (
	"context"
	"os"
	"testing"

	"github.com/sonpham/vielish/server/internal/database"
)

func setupTestDB(t *testing.T) *Repository {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := database.NewPostgres(ctx, dbURL)
	if err != nil {
		t.Fatalf("connecting to test db: %v", err)
	}

	// Clean up users table before each test
	pool.Exec(ctx, "DELETE FROM users")

	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM users")
		pool.Close()
	})

	return NewRepository(pool)
}

func TestRepository_CreateAndGetByEmail(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	user, err := repo.Create(ctx, "test@example.com", "hashed_password", "Test User")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if user.ID == "" {
		t.Error("Create() returned empty ID")
	}
	if user.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", user.Email, "test@example.com")
	}
	if user.DisplayName != "Test User" {
		t.Errorf("DisplayName = %q, want %q", user.DisplayName, "Test User")
	}
	if user.Level != "beginner" {
		t.Errorf("Level = %q, want %q", user.Level, "beginner")
	}

	found, err := repo.GetByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("GetByEmail() error = %v", err)
	}
	if found.ID != user.ID {
		t.Errorf("GetByEmail().ID = %q, want %q", found.ID, user.ID)
	}
	if found.PasswordHash != "hashed_password" {
		t.Errorf("GetByEmail().PasswordHash = %q, want %q", found.PasswordHash, "hashed_password")
	}
}

func TestRepository_GetByEmail_NotFound(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
	if err != ErrUserNotFound {
		t.Errorf("GetByEmail() error = %v, want ErrUserNotFound", err)
	}
}

func TestRepository_Create_DuplicateEmail(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	_, err := repo.Create(ctx, "dup@example.com", "hash1", "User 1")
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	_, err = repo.Create(ctx, "dup@example.com", "hash2", "User 2")
	if err != ErrEmailExists {
		t.Errorf("second Create() error = %v, want ErrEmailExists", err)
	}
}

func TestRepository_GetByID(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	user, err := repo.Create(ctx, "byid@example.com", "hashed", "By ID")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found.Email != "byid@example.com" {
		t.Errorf("GetByID().Email = %q, want %q", found.Email, "byid@example.com")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/auth/ -run TestRepository -v`
Expected: FAIL — `NewRepository`, `ErrUserNotFound`, `ErrEmailExists` undefined

- [ ] **Step 3: Write repository implementation**

```go
// server/internal/auth/repository.go
package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailExists  = errors.New("email already exists")
)

// RepositoryInterface defines the contract for user data access.
// Service depends on this interface, not the concrete Repository.
type RepositoryInterface interface {
	Create(ctx context.Context, email, passwordHash, displayName string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, email, passwordHash, displayName string) (*User, error) {
	var user User
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, display_name)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, password_hash, display_name, level, created_at`,
		email, passwordHash, displayName,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.Level, &user.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, display_name, level, created_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.Level, &user.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	var user User
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, display_name, level, created_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.Level, &user.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/auth/ -run TestRepository -v`
Expected: PASS (requires Docker postgres to be running with migration applied)

---

### Task 8: Auth Service

**Files:**
- Create: `server/internal/auth/service.go`
- Create: `server/internal/auth/service_test.go`

- [ ] **Step 1: Write service test**

```go
// server/internal/auth/service_test.go
package auth

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sonpham/vielish/server/internal/database"
)

func setupTestService(t *testing.T) *Service {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable"
	}
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	ctx := context.Background()

	pool, err := database.NewPostgres(ctx, dbURL)
	if err != nil {
		t.Fatalf("connecting to test db: %v", err)
	}

	rdb, err := database.NewRedis(ctx, redisURL)
	if err != nil {
		t.Fatalf("connecting to test redis: %v", err)
	}

	pool.Exec(ctx, "DELETE FROM users")

	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM users")
		pool.Close()
		rdb.FlushDB(context.Background())
		rdb.Close()
	})

	repo := NewRepository(pool)
	return NewService(repo, rdb, "test-jwt-secret")
}

func TestService_Register(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	tokens, err := svc.Register(ctx, "new@example.com", "password123", "New User")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if tokens.AccessToken == "" {
		t.Error("Register() returned empty access token")
	}
	if tokens.RefreshToken == "" {
		t.Error("Register() returned empty refresh token")
	}
	if tokens.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn = %d, want 3600", tokens.ExpiresIn)
	}
}

func TestService_Register_DuplicateEmail(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	_, err := svc.Register(ctx, "dup@example.com", "password123", "User 1")
	if err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	_, err = svc.Register(ctx, "dup@example.com", "password456", "User 2")
	if err != ErrEmailExists {
		t.Errorf("second Register() error = %v, want ErrEmailExists", err)
	}
}

func TestService_Login(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	_, err := svc.Register(ctx, "login@example.com", "password123", "Login User")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	tokens, err := svc.Login(ctx, "login@example.com", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if tokens.AccessToken == "" {
		t.Error("Login() returned empty access token")
	}
}

func TestService_Login_WrongPassword(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	_, err := svc.Register(ctx, "wrong@example.com", "password123", "User")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	_, err = svc.Login(ctx, "wrong@example.com", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Errorf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestService_Login_UserNotFound(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	_, err := svc.Login(ctx, "ghost@example.com", "password123")
	if err != ErrInvalidCredentials {
		t.Errorf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestService_ValidateToken(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	tokens, err := svc.Register(ctx, "validate@example.com", "password123", "Validate User")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	userID, err := svc.ValidateToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if userID == "" {
		t.Error("ValidateToken() returned empty userID")
	}
}

func TestService_RefreshToken(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	tokens, err := svc.Register(ctx, "refresh@example.com", "password123", "Refresh User")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Small delay to ensure different token
	time.Sleep(10 * time.Millisecond)

	newTokens, err := svc.Refresh(ctx, tokens.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if newTokens.AccessToken == "" {
		t.Error("Refresh() returned empty access token")
	}
	if newTokens.RefreshToken == "" {
		t.Error("Refresh() returned empty refresh token")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/auth/ -run TestService -v`
Expected: FAIL — `NewService`, `ErrInvalidCredentials` undefined

- [ ] **Step 3: Install JWT dependency**

Run: `cd server && go get github.com/golang-jwt/jwt/v5 golang.org/x/crypto`

- [ ] **Step 4: Write service implementation**

```go
// server/internal/auth/service.go
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

const (
	accessTokenExpiry  = time.Hour
	refreshTokenExpiry = 7 * 24 * time.Hour
	refreshTokenPrefix = "refresh:"
)

type Service struct {
	repo      RepositoryInterface
	redis     *redis.Client
	jwtSecret []byte
}

func NewService(repo RepositoryInterface, redis *redis.Client, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		redis:     redis,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *Service) Register(ctx context.Context, email, password, displayName string) (*TokenResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user, err := s.repo.Create(ctx, email, string(hash), displayName)
	if err != nil {
		return nil, err
	}

	return s.generateTokens(ctx, user.ID)
}

func (s *Service) Login(ctx context.Context, email, password string) (*TokenResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if errors.Is(err, ErrUserNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokens(ctx, user.ID)
}

func (s *Service) ValidateToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return "", fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token claims")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", errors.New("invalid subject claim")
	}

	return userID, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	key := refreshTokenPrefix + refreshToken
	userID, err := s.redis.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, errors.New("invalid refresh token")
	}
	if err != nil {
		return nil, fmt.Errorf("checking refresh token: %w", err)
	}

	// Delete old refresh token
	s.redis.Del(ctx, key)

	return s.generateTokens(ctx, userID)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	key := refreshTokenPrefix + refreshToken
	return s.redis.Del(ctx, key).Err()
}

func (s *Service) generateTokens(ctx context.Context, userID string) (*TokenResponse, error) {
	now := time.Now()

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(accessTokenExpiry).Unix(),
	})

	accessStr, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}
	refreshStr := hex.EncodeToString(refreshBytes)

	// Store refresh token in Redis
	err = s.redis.Set(ctx, refreshTokenPrefix+refreshStr, userID, refreshTokenExpiry).Err()
	if err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresIn:    int(accessTokenExpiry.Seconds()),
	}, nil
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd server && go test ./internal/auth/ -run TestService -v`
Expected: PASS

---

### Task 9: Auth Middleware

**Files:**
- Create: `server/internal/auth/middleware.go`
- Create: `server/internal/auth/middleware_test.go`

- [ ] **Step 1: Write middleware test**

```go
// server/internal/auth/middleware_test.go
package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func makeTestToken(secret string, userID string, expiry time.Duration) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(expiry).Unix(),
	})
	str, _ := token.SignedString([]byte(secret))
	return str
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	token := makeTestToken(secret, "user-123", time.Hour)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	var capturedUserID string
	r.Use(AuthMiddleware(secret))
	r.GET("/test", func(c *gin.Context) {
		capturedUserID = c.GetString("userID")
		c.Status(http.StatusOK)
	})

	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if capturedUserID != "user-123" {
		t.Errorf("userID = %q, want %q", capturedUserID, "user-123")
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(AuthMiddleware("test-secret"))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	token := makeTestToken(secret, "user-123", -time.Hour) // expired

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(AuthMiddleware(secret))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(AuthMiddleware("test-secret"))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/auth/ -run TestAuthMiddleware -v`
Expected: FAIL — `AuthMiddleware` undefined

- [ ] **Step 3: Write middleware implementation**

```go
// server/internal/auth/middleware.go
package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
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

		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid subject claim"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/auth/ -run TestAuthMiddleware -v`
Expected: PASS

---

### Task 10: Auth Handler

**Files:**
- Create: `server/internal/auth/handler.go`
- Create: `server/internal/auth/handler_test.go`

- [ ] **Step 1: Write handler test**

```go
// server/internal/auth/handler_test.go
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sonpham/vielish/server/internal/database"
)

func setupTestHandler(t *testing.T) (*Handler, *gin.Engine) {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable"
	}
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	ctx := context.Background()

	pool, err := database.NewPostgres(ctx, dbURL)
	if err != nil {
		t.Fatalf("connecting to test db: %v", err)
	}

	rdb, err := database.NewRedis(ctx, redisURL)
	if err != nil {
		t.Fatalf("connecting to test redis: %v", err)
	}

	pool.Exec(ctx, "DELETE FROM users")

	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM users")
		pool.Close()
		rdb.FlushDB(context.Background())
		rdb.Close()
	})

	repo := NewRepository(pool)
	svc := NewService(repo, rdb, "test-jwt-secret")
	handler := NewHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/auth/register", handler.Register)
	r.POST("/api/auth/login", handler.Login)
	r.POST("/api/auth/refresh", handler.Refresh)

	return handler, r
}

func TestHandler_Register_Success(t *testing.T) {
	_, r := setupTestHandler(t)

	body, _ := json.Marshal(RegisterRequest{
		Email:       "handler@example.com",
		Password:    "password123",
		DisplayName: "Handler Test",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d; body = %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var resp TokenResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.AccessToken == "" {
		t.Error("response missing access_token")
	}
	if resp.RefreshToken == "" {
		t.Error("response missing refresh_token")
	}
}

func TestHandler_Register_InvalidInput(t *testing.T) {
	_, r := setupTestHandler(t)

	body, _ := json.Marshal(map[string]string{
		"email": "not-an-email",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_Login_Success(t *testing.T) {
	_, r := setupTestHandler(t)

	// Register first
	regBody, _ := json.Marshal(RegisterRequest{
		Email:       "login-handler@example.com",
		Password:    "password123",
		DisplayName: "Login Test",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Now login
	loginBody, _ := json.Marshal(LoginRequest{
		Email:    "login-handler@example.com",
		Password: "password123",
	})
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp TokenResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.AccessToken == "" {
		t.Error("response missing access_token")
	}
}

func TestHandler_Login_WrongPassword(t *testing.T) {
	_, r := setupTestHandler(t)

	// Register first
	regBody, _ := json.Marshal(RegisterRequest{
		Email:       "wrong-handler@example.com",
		Password:    "password123",
		DisplayName: "Wrong Test",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Login with wrong password
	loginBody, _ := json.Marshal(LoginRequest{
		Email:    "wrong-handler@example.com",
		Password: "wrongpassword",
	})
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_Refresh_Success(t *testing.T) {
	_, r := setupTestHandler(t)

	// Register to get tokens
	regBody, _ := json.Marshal(RegisterRequest{
		Email:       "refresh-handler@example.com",
		Password:    "password123",
		DisplayName: "Refresh Test",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var regResp TokenResponse
	json.Unmarshal(w.Body.Bytes(), &regResp)

	// Refresh
	refreshBody, _ := json.Marshal(RefreshRequest{
		RefreshToken: regResp.RefreshToken,
	})
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewReader(refreshBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp TokenResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.AccessToken == "" {
		t.Error("response missing access_token")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/auth/ -run TestHandler -v`
Expected: FAIL — `NewHandler`, `Handler` undefined

- [ ] **Step 3: Write handler implementation**

```go
// server/internal/auth/handler.go
package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sonpham/vielish/server/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}

	tokens, err := h.service.Register(c.Request.Context(), req.Email, req.Password, req.DisplayName)
	if errors.Is(err, ErrEmailExists) {
		response.Error(c, http.StatusConflict, "email already registered")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "registration failed")
		return
	}

	response.Success(c, http.StatusCreated, tokens)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}

	tokens, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		response.Error(c, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "login failed")
		return
	}

	response.Success(c, http.StatusOK, tokens)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}

	tokens, err := h.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	response.Success(c, http.StatusOK, tokens)
}

func (h *Handler) Logout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}

	if err := h.service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.Error(c, http.StatusInternalServerError, "logout failed")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "logged out"})
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/auth/ -run TestHandler -v`
Expected: PASS

---

### Task 11: Router + Server Entrypoint

**Files:**
- Create: `server/internal/router/router.go`
- Create: `server/cmd/api/main.go`

- [ ] **Step 1: Write router**

Note: This initial router will be replaced with the CORS-enabled version in Task 18. Write it without CORS first to verify basic routing works.

```go
// server/internal/router/router.go
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sonpham/vielish/server/internal/auth"
)

func New(authHandler *auth.Handler, jwtSecret string, corsOrigins []string) *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes (public)
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.POST("/logout", authHandler.Logout)
	}

	// Protected routes (placeholder for future features)
	_ = r.Group("/api").Use(auth.AuthMiddleware(jwtSecret))

	return r
}
```

- [ ] **Step 2: Write server entrypoint**

```go
// server/cmd/api/main.go
package main

import (
	"context"
	"log"

	"github.com/sonpham/vielish/server/internal/auth"
	"github.com/sonpham/vielish/server/internal/config"
	"github.com/sonpham/vielish/server/internal/database"
	"github.com/sonpham/vielish/server/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	ctx := context.Background()

	db, err := database.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	rdb, err := database.NewRedis(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, rdb, cfg.JWTSecret)
	authHandler := auth.NewHandler(authService)

	r := router.New(authHandler, cfg.JWTSecret, cfg.CORSOrigins)

	log.Printf("Starting server on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

- [ ] **Step 3: Verify compilation**

Run: `cd server && go build ./cmd/api/`
Expected: No errors

- [ ] **Step 4: Manual smoke test**

Run (in separate terminal): `cd server && go run cmd/api/main.go`
Expected: "Starting server on :8080"

Run: `curl -s http://localhost:8080/api/health | jq .`
Expected: `{"status": "ok"}`

Run: `curl -s -X POST http://localhost:8080/api/auth/register -H "Content-Type: application/json" -d '{"email":"test@test.com","password":"password123","display_name":"Test"}' | jq .`
Expected: JSON with access_token, refresh_token, expires_in

---

### Task 12: Next.js Project Setup

**Files:**
- Create: `web/` (via create-next-app)
- Modify: `web/package.json`

- [ ] **Step 1: Create Next.js project**

Run: `npx create-next-app@latest web --typescript --tailwind --eslint --app --src-dir=false --import-alias="@/*" --no-turbopack`

- [ ] **Step 2: Verify project runs**

Run: `cd web && npm run dev`
Expected: Next.js dev server starts on port 3000

- [ ] **Step 3: Clean up default content**

Replace `web/app/page.tsx` with a minimal landing page:

```tsx
// web/app/page.tsx
export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8">
      <h1 className="text-4xl font-bold mb-4">Vielish</h1>
      <p className="text-lg text-gray-600 mb-8">
        Learn English the Vietnamese way
      </p>
      <div className="flex gap-4">
        <a
          href="/login"
          className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          Đăng nhập
        </a>
        <a
          href="/register"
          className="px-6 py-3 border border-blue-600 text-blue-600 rounded-lg hover:bg-blue-50"
        >
          Đăng ký
        </a>
      </div>
    </main>
  );
}
```

---

### Task 13: API Client

**Files:**
- Create: `web/lib/api.ts`

- [ ] **Step 1: Write API client with token handling**

```typescript
// web/lib/api.ts
const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface TokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

class ApiClient {
  private getAccessToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("access_token");
  }

  private getRefreshToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("refresh_token");
  }

  private saveTokens(tokens: TokenResponse): void {
    localStorage.setItem("access_token", tokens.access_token);
    localStorage.setItem("refresh_token", tokens.refresh_token);
  }

  clearTokens(): void {
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
  }

  async request(path: string, options: RequestInit = {}): Promise<Response> {
    const token = this.getAccessToken();
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...((options.headers as Record<string, string>) || {}),
    };

    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    let response = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers,
    });

    // Try refreshing token on 401
    if (response.status === 401 && this.getRefreshToken()) {
      const refreshed = await this.refresh();
      if (refreshed) {
        headers["Authorization"] = `Bearer ${this.getAccessToken()}`;
        response = await fetch(`${API_BASE}${path}`, {
          ...options,
          headers,
        });
      }
    }

    return response;
  }

  async register(
    email: string,
    password: string,
    displayName: string
  ): Promise<TokenResponse> {
    const res = await fetch(`${API_BASE}/api/auth/register`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password, display_name: displayName }),
    });

    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || "Registration failed");
    }

    const tokens: TokenResponse = await res.json();
    this.saveTokens(tokens);
    return tokens;
  }

  async login(email: string, password: string): Promise<TokenResponse> {
    const res = await fetch(`${API_BASE}/api/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });

    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || "Login failed");
    }

    const tokens: TokenResponse = await res.json();
    this.saveTokens(tokens);
    return tokens;
  }

  private async refresh(): Promise<boolean> {
    const refreshToken = this.getRefreshToken();
    if (!refreshToken) return false;

    try {
      const res = await fetch(`${API_BASE}/api/auth/refresh`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      if (!res.ok) {
        this.clearTokens();
        return false;
      }

      const tokens: TokenResponse = await res.json();
      this.saveTokens(tokens);
      return true;
    } catch {
      this.clearTokens();
      return false;
    }
  }
}

export const api = new ApiClient();
```

---

### Task 14: Auth Context

**Files:**
- Create: `web/lib/auth-context.tsx`

- [ ] **Step 1: Write auth context provider**

```tsx
// web/lib/auth-context.tsx
"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  ReactNode,
} from "react";
import { api } from "./api";

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (
    email: string,
    password: string,
    displayName: string
  ) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem("access_token");
    setIsAuthenticated(!!token);
    setIsLoading(false);
  }, []);

  const login = async (email: string, password: string) => {
    await api.login(email, password);
    setIsAuthenticated(true);
  };

  const register = async (
    email: string,
    password: string,
    displayName: string
  ) => {
    await api.register(email, password, displayName);
    setIsAuthenticated(true);
  };

  const logout = () => {
    api.clearTokens();
    setIsAuthenticated(false);
  };

  return (
    <AuthContext.Provider
      value={{ isAuthenticated, isLoading, login, register, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
```

---

### Task 15: Auth Form Component

**Files:**
- Create: `web/components/auth-form.tsx`

- [ ] **Step 1: Write reusable auth form**

```tsx
// web/components/auth-form.tsx
"use client";

import { useState, FormEvent } from "react";

interface AuthFormProps {
  mode: "login" | "register";
  onSubmit: (data: {
    email: string;
    password: string;
    displayName?: string;
  }) => Promise<void>;
}

export default function AuthForm({ mode, onSubmit }: AuthFormProps) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await onSubmit({
        email,
        password,
        ...(mode === "register" ? { displayName } : {}),
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="w-full max-w-sm space-y-4">
      {error && (
        <div className="p-3 bg-red-50 text-red-700 rounded-lg text-sm">
          {error}
        </div>
      )}

      {mode === "register" && (
        <div>
          <label
            htmlFor="displayName"
            className="block text-sm font-medium mb-1"
          >
            Tên hiển thị
          </label>
          <input
            id="displayName"
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            required
            className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
      )}

      <div>
        <label htmlFor="email" className="block text-sm font-medium mb-1">
          Email (Địa chỉ email)
        </label>
        <input
          id="email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      <div>
        <label htmlFor="password" className="block text-sm font-medium mb-1">
          Mật khẩu
        </label>
        <input
          id="password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          minLength={8}
          className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      <button
        type="submit"
        disabled={loading}
        className="w-full py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
      >
        {loading
          ? "Đang tải..."
          : mode === "login"
            ? "Đăng nhập"
            : "Tạo tài khoản"}
      </button>
    </form>
  );
}
```

---

### Task 16: Login + Register Pages

**Files:**
- Create: `web/app/login/page.tsx`
- Create: `web/app/register/page.tsx`
- Modify: `web/app/layout.tsx`

- [ ] **Step 1: Update root layout to include AuthProvider**

```tsx
// web/app/layout.tsx
import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "@/lib/auth-context";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Vielish - Learn English",
  description: "English learning app for Vietnamese speakers",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="vi">
      <body className={inter.className}>
        <AuthProvider>{children}</AuthProvider>
      </body>
    </html>
  );
}
```

- [ ] **Step 2: Write login page**

```tsx
// web/app/login/page.tsx
"use client";

import { useRouter } from "next/navigation";
import AuthForm from "@/components/auth-form";
import { useAuth } from "@/lib/auth-context";

export default function LoginPage() {
  const router = useRouter();
  const { login } = useAuth();

  const handleLogin = async (data: {
    email: string;
    password: string;
  }) => {
    await login(data.email, data.password);
    router.push("/dashboard");
  };

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8">
      <h1 className="text-3xl font-bold mb-8">Đăng nhập Vielish</h1>
      <AuthForm mode="login" onSubmit={handleLogin} />
      <p className="mt-4 text-sm text-gray-600">
        Chưa có tài khoản?{" "}
        <a href="/register" className="text-blue-600 hover:underline">
          Đăng ký
        </a>
      </p>
    </main>
  );
}
```

- [ ] **Step 3: Write register page**

```tsx
// web/app/register/page.tsx
"use client";

import { useRouter } from "next/navigation";
import AuthForm from "@/components/auth-form";
import { useAuth } from "@/lib/auth-context";

export default function RegisterPage() {
  const router = useRouter();
  const { register } = useAuth();

  const handleRegister = async (data: {
    email: string;
    password: string;
    displayName?: string;
  }) => {
    await register(data.email, data.password, data.displayName || "");
    router.push("/dashboard");
  };

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8">
      <h1 className="text-3xl font-bold mb-8">Tạo tài khoản</h1>
      <AuthForm mode="register" onSubmit={handleRegister} />
      <p className="mt-4 text-sm text-gray-600">
        Đã có tài khoản?{" "}
        <a href="/login" className="text-blue-600 hover:underline">
          Đăng nhập
        </a>
      </p>
    </main>
  );
}
```

---

### Task 17: Protected Dashboard Page

**Files:**
- Create: `web/app/dashboard/layout.tsx`
- Create: `web/app/dashboard/page.tsx`

- [ ] **Step 1: Write dashboard layout with auth guard**

```tsx
// web/app/dashboard/layout.tsx
"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, isLoading, logout } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push("/login");
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-gray-500">Đang tải...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen">
      <nav className="border-b px-6 py-4 flex justify-between items-center">
        <h1 className="text-xl font-bold">Vielish</h1>
        <button
          onClick={() => {
            logout();
            router.push("/login");
          }}
          className="text-sm text-gray-600 hover:text-gray-900"
        >
          Đăng xuất
        </button>
      </nav>
      <main className="p-6">{children}</main>
    </div>
  );
}
```

- [ ] **Step 2: Write dashboard page (placeholder)**

```tsx
// web/app/dashboard/page.tsx
export default function DashboardPage() {
  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Bảng điều khiển</h2>
      <p className="text-gray-600">
        Chào mừng bạn đến với Vielish! Bảng học tập của bạn sẽ hiển thị ở đây.
      </p>
    </div>
  );
}
```

---

### Task 18: CORS Configuration

**Files:**
- Modify: `server/internal/router/router.go`

- [ ] **Step 1: Install CORS middleware**

Run: `cd server && go get github.com/gin-contrib/cors`

- [ ] **Step 2: Update router with CORS**

```go
// server/internal/router/router.go
package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sonpham/vielish/server/internal/auth"
)

func New(authHandler *auth.Handler, jwtSecret string, corsOrigins []string) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes (public)
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.POST("/logout", authHandler.Logout)
	}

	// Protected routes (placeholder for future features)
	_ = r.Group("/api").Use(auth.AuthMiddleware(jwtSecret))

	return r
}
```

- [ ] **Step 3: Verify compilation**

Run: `cd server && go build ./cmd/api/`
Expected: No errors

---

### Task 19: Environment Example File

**Files:**
- Create: `.env.example`

- [ ] **Step 1: Create .env.example with all config variables**

```bash
# .env.example
# App
APP_ENV=development
PORT=8080

# Database
DATABASE_URL=postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# Auth
JWT_SECRET=dev-secret-change-in-production

# CORS (comma-separated origins)
CORS_ORIGINS=http://localhost:3000

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:8080
```

- [ ] **Step 2: Add .env to .gitignore**

Ensure `.env` is in `.gitignore` (not `.env.example`):

```
# .gitignore (append if not already present)
.env
```

---

### Task 20: End-to-End Verification

- [ ] **Step 1: Start all services**

Run: `docker-compose up -d`
Run: `migrate -path server/migrations -database "postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable" up`
Run (terminal 1): `cd server && go run cmd/api/main.go`
Run (terminal 2): `cd web && npm run dev`

- [ ] **Step 2: Test auth flow via curl**

```bash
# Register
curl -s -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"e2e@test.com","password":"password123","display_name":"E2E User"}' | jq .
# Expected: JSON with access_token, refresh_token, expires_in

# Login
curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"e2e@test.com","password":"password123"}' | jq .
# Expected: JSON with access_token, refresh_token, expires_in

# Logout (use the refresh_token from login response)
curl -s -X POST http://localhost:8080/api/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token_from_login>"}' | jq .
# Expected: {"message": "logged out"}

# Health (should work without auth)
curl -s http://localhost:8080/api/health | jq .
# Expected: {"status": "ok"}
```

- [ ] **Step 3: Test frontend**

Open browser to `http://localhost:3000`
Expected: Landing page with "Đăng nhập"/"Đăng ký" links

Navigate to `/register`, create account
Expected: Redirected to `/dashboard` after successful registration

Navigate to `/login`, login with created credentials
Expected: Redirected to `/dashboard` after successful login

- [ ] **Step 4: Run all backend tests**

Run: `cd server && go test ./... -v`
Expected: All tests PASS
