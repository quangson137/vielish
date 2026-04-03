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
