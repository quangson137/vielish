package driven

import (
	"time"

	domain "github.com/sonpham/vielish/server/internal/domain/auth"
)

type UserModel struct {
	ID           string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	DisplayName  string `gorm:"not null"`
	Level        string `gorm:"default:'beginner'"`
	CreatedAt    time.Time
}

func (UserModel) TableName() string { return "users" }

func (m *UserModel) ToEntity() *domain.User {
	return &domain.User{
		ID:           m.ID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		DisplayName:  m.DisplayName,
		Level:        m.Level,
		CreatedAt:    m.CreatedAt,
	}
}
