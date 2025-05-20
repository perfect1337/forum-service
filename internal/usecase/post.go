package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/perfect1337/forum-service/internal/entity"
)

type PostUseCase interface {
	CreatePost(ctx context.Context, post *entity.Post) error
	GetPostByID(ctx context.Context, id int) (*entity.Post, error)
	GetAllPosts(ctx context.Context) ([]*entity.Post, error)
	DeletePost(ctx context.Context, postID, userID int) error
	UpdatePost(ctx context.Context, postID int, userID int, title, content string) error
}

type PostRepository interface {
	CreatePost(ctx context.Context, post *entity.Post) error
	GetPostByID(ctx context.Context, id int) (*entity.Post, error)
	GetAllPosts(ctx context.Context) ([]*entity.Post, error)
	DeletePost(ctx context.Context, id int) error
	UpdatePost(ctx context.Context, postID int, title, content string) error
}

type UserRepository interface {
	GetUserByID(ctx context.Context, id int) (*entity.User, error)
}
type PostService struct {
	postRepo PostRepository
	userRepo UserRepository
}

type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
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

	fmt.Printf("Debug: post.UserID=%d, userID=%d, user.Role=%s\n", post.UserID, userID, user.Role)

	if post.UserID != userID && user.Role != "admin" {
		return errors.New("unauthorized: you can only delete your own posts")
	}

	return s.postRepo.DeletePost(ctx, postID)
}
func (s *PostService) CreatePost(ctx context.Context, post *entity.Post) error {
	if post == nil {
		return errors.New("post cannot be nil")
	}
	if post.Title == "" {
		return errors.New("post title cannot be empty")
	}
	if post.Content == "" {
		return errors.New("post content cannot be empty")
	}
	if post.UserID == 0 {
		return errors.New("user ID cannot be empty")
	}
	return s.postRepo.CreatePost(ctx, post)
}

func (s *PostService) GetPostByID(ctx context.Context, id int) (*entity.Post, error) {
	return s.postRepo.GetPostByID(ctx, id)
}

func (s *PostService) GetAllPosts(ctx context.Context) ([]*entity.Post, error) {
	return s.postRepo.GetAllPosts(ctx)
}

func (s *PostService) UpdatePost(ctx context.Context, postID int, userID int, title, content string) error {
	post, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return err
	}
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if post.UserID != userID && user.Role != "admin" {
		return errors.New("unauthorized: you can only update your own posts")
	}
	return s.postRepo.UpdatePost(ctx, postID, title, content)
}

func NewPostUseCase(postRepo PostRepository, userRepo UserRepository) PostUseCase {
	return &PostService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}
