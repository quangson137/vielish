package driven

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	domain "github.com/sonpham/vielish/server/internal/domain/auth"
)

const refreshTokenPrefix = "refresh:"

type Repository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewRepository(db *gorm.DB, redis *redis.Client) *Repository {
	return &Repository{db: db, redis: redis}
}

func (r *Repository) Create(ctx context.Context, email, passwordHash, displayName string) (*domain.User, error) {
	m := &UserModel{Email: email, PasswordHash: passwordHash, DisplayName: displayName}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrEmailExists
		}
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return m.ToEntity(), nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by email: %w", err)
	}
	return m.ToEntity(), nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by id: %w", err)
	}
	return m.ToEntity(), nil
}

func (r *Repository) StoreRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	return r.redis.Set(ctx, refreshTokenPrefix+token, userID, ttl).Err()
}

func (r *Repository) GetRefreshToken(ctx context.Context, token string) (string, error) {
	userID, err := r.redis.Get(ctx, refreshTokenPrefix+token).Result()
	if errors.Is(err, redis.Nil) {
		return "", domain.ErrInvalidToken
	}
	if err != nil {
		return "", fmt.Errorf("getting refresh token: %w", err)
	}
	return userID, nil
}

func (r *Repository) DeleteRefreshToken(ctx context.Context, token string) error {
	return r.redis.Del(ctx, refreshTokenPrefix+token).Err()
}
