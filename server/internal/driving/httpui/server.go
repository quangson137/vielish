package httpui

import (
	"context"
	"errors"
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

func RegisterLifecycle(lc fx.Lifecycle, shutter fx.Shutdowner, r *gin.Engine, cfg config.Config, log *zap.Logger) {
	srv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: r,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("starting HTTP server", zap.String("port", cfg.App.Port))
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Error("HTTP server error", zap.Error(err))
					_ = shutter.Shutdown(fx.ExitCode(1))
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
