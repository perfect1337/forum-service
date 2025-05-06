package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/perfect1337/forum-service/internal/config"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/repository"
	"github.com/stretchr/testify/mock"
)

// Интерфейс UseCase
type PostUseCase interface {
	CreatePost(ctx context.Context, post *entity.Post) error
	GetPostByID(ctx context.Context, id int) (*entity.Post, error)
	GetAllPosts(ctx context.Context) ([]*entity.Post, error)
	DeletePost(ctx context.Context, postID, userID int) error // Два параметра!
}

type PostRepository interface {
	CreatePost(ctx context.Context, post *entity.Post) error
	GetPostByID(ctx context.Context, id int) (*entity.Post, error)
	GetAllPosts(ctx context.Context) ([]*entity.Post, error)
	DeletePost(ctx context.Context, id int) error
}

type UserRepository interface {
	GetUserByID(ctx context.Context, id int) (*entity.User, error)
}
type MockPostUseCase struct {
	mock.Mock
}

// Реализация интерфейса
type PostService struct {
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

// Реализация методов PostUseCase
func (s *PostService) DeletePost(ctx context.Context, postID, userID int) error {
	post, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if post.UserID != userID && user.Role != "admin" {
		return errors.New("unauthorized: you can only delete your own posts")
	}

	return s.postRepo.DeletePost(ctx, postID)
}

func (s *PostService) CreatePost(ctx context.Context, post *entity.Post) error {
	return s.postRepo.CreatePost(ctx, post)
}

func (s *PostService) GetPostByID(ctx context.Context, id int) (*entity.Post, error) {
	return s.postRepo.GetPostByID(ctx, id)
}

func (s *PostService) GetAllPosts(ctx context.Context) ([]*entity.Post, error) {
	return s.postRepo.GetAllPosts(ctx)
}

// Конструкторы
func NewAuthUseCase(repo *repository.Postgres, cfg *config.Config) *AuthUseCase {
	return &AuthUseCase{
		repo:      repo,
		cfg:       cfg,
		SecretKey: []byte(cfg.Auth.SecretKey),
	}
}

func NewPostUseCase(postRepo PostRepository, userRepo UserRepository) PostUseCase {
	return &PostService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}
