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

func TestHandler_Login_200(t *testing.T) {
	stub := &stubUseCase{
		loginFn: func(_ context.Context, _ appcore.LoginInput) (*appcore.TokenOutput, error) {
			return &appcore.TokenOutput{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600}, nil
		},
	}
	_, r := newTestHandler(stub)
	w := postJSON(t, r, "/login", map[string]string{"email": "a@b.com", "password": "pass1234"})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200; body: %s", w.Code, w.Body)
	}
}

func TestHandler_Register_400_InvalidInput(t *testing.T) {
	stub := &stubUseCase{
		registerFn: func(_ context.Context, _ appcore.RegisterInput) (*appcore.TokenOutput, error) {
			return &appcore.TokenOutput{}, nil
		},
	}
	_, r := newTestHandler(stub)
	// missing display_name, short password
	w := postJSON(t, r, "/register", map[string]string{"email": "bad-email", "password": "short"})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400; body: %s", w.Code, w.Body)
	}
}

func TestHandler_Refresh_200(t *testing.T) {
	stub := &stubUseCase{
		refreshFn: func(_ context.Context, _ string) (*appcore.TokenOutput, error) {
			return &appcore.TokenOutput{AccessToken: "new-at", RefreshToken: "new-rt", ExpiresIn: 3600}, nil
		},
	}
	_, r := newTestHandler(stub)
	w := postJSON(t, r, "/refresh", map[string]string{"refresh_token": "old-rt"})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200; body: %s", w.Code, w.Body)
	}
}

func TestHandler_Refresh_401_InvalidToken(t *testing.T) {
	stub := &stubUseCase{
		refreshFn: func(_ context.Context, _ string) (*appcore.TokenOutput, error) {
			return nil, domain.ErrInvalidToken
		},
	}
	_, r := newTestHandler(stub)
	w := postJSON(t, r, "/refresh", map[string]string{"refresh_token": "bad-rt"})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401; body: %s", w.Code, w.Body)
	}
}

func TestHandler_Logout_200(t *testing.T) {
	stub := &stubUseCase{
		logoutFn: func(_ context.Context, _ string) error { return nil },
	}
	_, r := newTestHandler(stub)
	w := postJSON(t, r, "/logout", map[string]string{"refresh_token": "rt"})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200; body: %s", w.Code, w.Body)
	}
}

// keep compiler happy
var _ = errors.New
