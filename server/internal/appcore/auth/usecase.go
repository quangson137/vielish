package appcore

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/sonpham/vielish/server/internal/domain/auth"
	"github.com/sonpham/vielish/server/pkg/config"
)

type UseCase struct {
	repo       domain.Repository
	service    *domain.Service
	refreshTTL time.Duration
}

func NewUseCase(repo domain.Repository, service *domain.Service, cfg config.Config) *UseCase {
	return &UseCase{
		repo:       repo,
		service:    service,
		refreshTTL: cfg.JWT.RefreshTTL,
	}
}

func (uc *UseCase) Register(ctx context.Context, input RegisterInput) (*TokenOutput, error) {
	hash, err := uc.service.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}
	user, err := uc.repo.Create(ctx, input.Email, hash, input.DisplayName)
	if err != nil {
		return nil, err
	}
	return uc.issueTokens(ctx, user.ID)
}

func (uc *UseCase) Login(ctx context.Context, input LoginInput) (*TokenOutput, error) {
	user, err := uc.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if err := uc.service.CheckPassword(user.PasswordHash, input.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	return uc.issueTokens(ctx, user.ID)
}

func (uc *UseCase) Refresh(ctx context.Context, refreshToken string) (*TokenOutput, error) {
	userID, err := uc.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	if err := uc.repo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("deleting old refresh token: %w", err)
	}
	return uc.issueTokens(ctx, userID)
}

func (uc *UseCase) Logout(ctx context.Context, refreshToken string) error {
	return uc.repo.DeleteRefreshToken(ctx, refreshToken)
}

func (uc *UseCase) issueTokens(ctx context.Context, userID string) (*TokenOutput, error) {
	accessToken, err := uc.service.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}
	refreshToken, err := uc.service.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	if err := uc.repo.StoreRefreshToken(ctx, refreshToken, userID, uc.refreshTTL); err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}
	return &TokenOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    uc.service.AccessTTLSeconds(),
	}, nil
}
