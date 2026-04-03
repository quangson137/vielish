package database

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/sonpham/vielish/server/pkg/config"
)

func NewGorm(cfg config.Config, log *zap.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("opening gorm db: %w", err)
	}
	log.Info("connected to postgres")
	return db, nil
}
