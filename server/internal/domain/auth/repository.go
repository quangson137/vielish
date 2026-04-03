package domain

import (
	"context"
	"time"
)

// Repository is the only interface the domain owns for user persistence.
// Refresh token operations are included here for MVP simplicity.
type Repository interface {
	Create(ctx context.Context, email, passwordHash, displayName string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	StoreRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error
	GetRefreshToken(ctx context.Context, token string) (string, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}
