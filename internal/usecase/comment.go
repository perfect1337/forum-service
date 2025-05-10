package usecase

import (
	"context"
	"errors"

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
type CommentUseCaseInterface interface {
	CreateComment(ctx context.Context, comment *entity.Comment) error
	GetCommentsByPostID(ctx context.Context, postID int) ([]entity.Comment, error)
	DeleteComment(ctx context.Context, commentID, userID int) error
}

func NewCommentUseCase(repo CommentRepository) *CommentUseCase {
	return &CommentUseCase{repo: repo}
}
func (uc *CommentUseCase) CreateComment(ctx context.Context, comment *entity.Comment) error {
	if comment == nil {
		return errors.New("comment cannot be nil")
	}
	if comment.Content == "" {
		return errors.New("comment content cannot be empty")
	}
	if comment.PostID == 0 {
		return errors.New("post ID cannot be empty")
	}
	if comment.UserID == 0 {
		return errors.New("user ID cannot be empty")
	}
	return uc.repo.CreateComment(ctx, comment)
}

func (uc *CommentUseCase) GetCommentsByPostID(ctx context.Context, postID int) ([]entity.Comment, error) {
	if postID <= 0 {
		return nil, errors.New("invalid post ID")
	}
	return uc.repo.GetCommentsByPostID(ctx, postID)
}
func (uc *CommentUseCase) DeleteComment(ctx context.Context, commentID int, userID int) error {
	if commentID <= 0 {
		return errors.New("invalid comment ID")
	}
	if userID <= 0 {
		return errors.New("invalid user ID")
	}
	return uc.repo.DeleteComment(ctx, commentID, userID)
}
