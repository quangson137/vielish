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
