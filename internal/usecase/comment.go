package usecase

import (
	"context"

	"github.com/perfect1337/forum-service/internal/entity"
)

type CommentUseCase struct {
	repo CommentRepository
}
type CommentRepository interface {
	CreateComment(ctx context.Context, comment *entity.Comment) error
	GetCommentsByPostID(ctx context.Context, postID int) ([]entity.Comment, error)
	DeleteComment(ctx context.Context, commentID int, userID int) error
}

func NewCommentUseCase(repo CommentRepository) *CommentUseCase {
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
