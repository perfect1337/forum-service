package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/perfect1337/forum-service/internal/config"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/repository"
)

type PostUseCase struct {
	repo *repository.Postgres
}

type AuthUseCase struct {
	repo      *repository.Postgres
	cfg       *config.Config
	SecretKey []byte
}
type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

func (uc *AuthUseCase) ParseToken(tokenString string) (int64, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return uc.SecretKey, nil
	})

	if err != nil {
		return 0, "", err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims.UserID, claims.Username, nil
	}

	return 0, "", errors.New("invalid token claims")
}

// internal/usecase/post.go
func (uc *PostUseCase) DeletePost(ctx context.Context, postID int) error {
	return uc.repo.DeletePost(ctx, postID)
}
func (p *PostUseCase) CreatePost(ctx context.Context, post *entity.Post) error {
	return p.repo.CreatePost(ctx, post)
}

func (p *PostUseCase) GetPostByID(ctx context.Context, id int) (*entity.Post, error) {
	return p.repo.GetPostByID(ctx, id)
}

func (p *PostUseCase) GetAllPosts(ctx context.Context) ([]*entity.Post, error) {
	return p.repo.GetAllPosts(ctx)
}
func NewAuthUseCase(repo *repository.Postgres, cfg *config.Config) *AuthUseCase {
	return &AuthUseCase{repo: repo, cfg: cfg, SecretKey: []byte(cfg.Auth.SecretKey)}
}

func NewPostUseCase(repo *repository.Postgres) *PostUseCase {
	return &PostUseCase{repo: repo}
}
