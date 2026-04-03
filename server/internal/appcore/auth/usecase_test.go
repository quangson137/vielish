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
