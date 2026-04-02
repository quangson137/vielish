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
