package auth

import (
	"context"
	"os"
	"testing"
	"time"

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
