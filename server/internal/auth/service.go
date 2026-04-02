package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

const (
	accessTokenExpiry  = time.Hour
	refreshTokenExpiry = 7 * 24 * time.Hour
	refreshTokenPrefix = "refresh:"
)

type Service struct {
	repo      RepositoryInterface
	redis     *redis.Client
	jwtSecret []byte
}

func NewService(repo RepositoryInterface, redis *redis.Client, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		redis:     redis,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *Service) Register(ctx context.Context, email, password, displayName string) (*TokenResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user, err := s.repo.Create(ctx, email, string(hash), displayName)
	if err != nil {
		return nil, err
	}

	return s.generateTokens(ctx, user.ID)
}

func (s *Service) Login(ctx context.Context, email, password string) (*TokenResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if errors.Is(err, ErrUserNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokens(ctx, user.ID)
}

func (s *Service) ValidateToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return "", fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token claims")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", errors.New("invalid subject claim")
	}

	return userID, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	key := refreshTokenPrefix + refreshToken
	userID, err := s.redis.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, errors.New("invalid refresh token")
	}
	if err != nil {
		return nil, fmt.Errorf("checking refresh token: %w", err)
	}

	// Delete old refresh token
	s.redis.Del(ctx, key)

	return s.generateTokens(ctx, userID)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	key := refreshTokenPrefix + refreshToken
	return s.redis.Del(ctx, key).Err()
}

func (s *Service) generateTokens(ctx context.Context, userID string) (*TokenResponse, error) {
	now := time.Now()

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(accessTokenExpiry).Unix(),
	})

	accessStr, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}
	refreshStr := hex.EncodeToString(refreshBytes)

	// Store refresh token in Redis
	err = s.redis.Set(ctx, refreshTokenPrefix+refreshStr, userID, refreshTokenExpiry).Err()
	if err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresIn:    int(accessTokenExpiry.Seconds()),
	}, nil
}
