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

type PostRepository interface {
	CreatePost(ctx context.Context, post *entity.Post) error
	GetPostByID(ctx context.Context, id int) (*entity.Post, error)
	GetAllPosts(ctx context.Context) ([]*entity.Post, error)
	DeletePost(ctx context.Context, id int) error
}

type UserRepository interface {
	GetUserByID(ctx context.Context, id int) (*entity.User, error)
}
type PostUseCase struct {
	postRepo PostRepository
	userRepo UserRepository
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

func (uc *PostUseCase) DeletePost(ctx context.Context, postID, userID int) error {
	post, err := uc.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return err
	}

	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if post.UserID != userID && user.Role != "admin" {
		return errors.New("unauthorized: you can only delete your own posts")
	}

	return uc.postRepo.DeletePost(ctx, postID)
}
func (p *PostUseCase) CreatePost(ctx context.Context, post *entity.Post) error {
	return p.postRepo.CreatePost(ctx, post)
}

func (p *PostUseCase) GetPostByID(ctx context.Context, id int) (*entity.Post, error) {
	return p.postRepo.GetPostByID(ctx, id)
}

func (p *PostUseCase) GetAllPosts(ctx context.Context) ([]*entity.Post, error) {
	return p.postRepo.GetAllPosts(ctx)
}
func NewAuthUseCase(repo *repository.Postgres, cfg *config.Config) *AuthUseCase {
	return &AuthUseCase{repo: repo, cfg: cfg, SecretKey: []byte(cfg.Auth.SecretKey)}
}

func NewPostUseCase(postRepo PostRepository, userRepo UserRepository) *PostUseCase {
	return &PostUseCase{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}
