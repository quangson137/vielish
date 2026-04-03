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
