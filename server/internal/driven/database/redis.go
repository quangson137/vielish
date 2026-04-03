package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/sonpham/vielish/server/pkg/config"
)

func NewRedis(lc fx.Lifecycle, cfg config.Config, log *zap.Logger) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis url: %w", err)
	}
	client := redis.NewClient(opt)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("closing redis connection")
			return client.Close()
		},
	})
	log.Info("connected to redis")
	return client, nil
}
