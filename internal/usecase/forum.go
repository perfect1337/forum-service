package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/perfect1337/forum-service/internal/config"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/repository"
)

type AuthUseCase struct {
	repo      repository.Postgres
	cfg       *config.Config
	secretKey []byte
}

func NewAuthUseCase(repo *repository.Postgres, cfg *config.Config) *AuthUseCase {
	return &AuthUseCase{repo: repo, cfg: cfg, secretKey: []byte(cfg.Auth.SecretKey)}
}

func (a *AuthUseCase) Register(ctx context.Context, user *entity.User) error {
	return a.repo.CreateUser(ctx, user)
}

func (a *AuthUseCase) Login(ctx context.Context, email, password string) (*entity.AuthResponse, error) {
	user, err := a.repo.GetUserByCredentials(ctx, email, password)
	if err != nil {
		return nil, err
	}

	accessToken, err := a.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := a.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &entity.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (a *AuthUseCase) RefreshTokens(ctx context.Context, refreshToken string) (*entity.AuthResponse, error) {
	rt, err := a.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	if rt.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("refresh token expired")
	}

	user, err := a.repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	accessToken, err := a.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := a.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	err = a.repo.DeleteRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	err = a.repo.CreateRefreshToken(ctx, &entity.RefreshToken{
		UserID:    user.ID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(a.cfg.Auth.RefreshTokenDuration),
	})
	if err != nil {
		return nil, err
	}

	return &entity.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         *user,
	}, nil
}

func (a *AuthUseCase) generateAccessToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(a.cfg.Auth.AccessTokenDuration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.secretKey)
}

func (a *AuthUseCase) generateRefreshToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(a.cfg.Auth.RefreshTokenDuration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.secretKey)
}
