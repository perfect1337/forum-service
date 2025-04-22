package usecase

import (
	"context"

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
