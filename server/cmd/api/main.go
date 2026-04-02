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
