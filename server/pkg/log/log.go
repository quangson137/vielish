package log

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/sonpham/vielish/server/pkg/config"
)

func NewLogger(lc fx.Lifecycle, cfg config.Config) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if cfg.App.Env == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			_ = logger.Sync()
			return nil
		},
	})
	return logger, nil
}

var Module = fx.Module("log",
	fx.Provide(NewLogger),
)
