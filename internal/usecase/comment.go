package usecase

import (
	"context"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/repository"
)

type CommentUseCase struct {
	repo *repository.Postgres
}

func NewCommentUseCase(repo *repository.Postgres) *CommentUseCase {
	return &CommentUseCase{repo: repo}
}

func (uc *CommentUseCase) CreateComment(ctx context.Context, comment *entity.Comment) error {
	return uc.repo.CreateComment(ctx, comment)
}

func (uc *CommentUseCase) GetCommentsByPostID(ctx context.Context, postID int) ([]entity.Comment, error) {
	return uc.repo.GetCommentsByPostID(ctx, postID)
}

func (uc *CommentUseCase) DeleteComment(ctx context.Context, commentID int, userID int) error {
	return uc.repo.DeleteComment(ctx, commentID, userID)
}
