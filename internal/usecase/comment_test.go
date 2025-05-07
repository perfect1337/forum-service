package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/mocks"
	"github.com/perfect1337/forum-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCommentUseCase_CreateComment(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := new(mocks.MockCommentRepository)
		uc := usecase.NewCommentUseCase(repo)

		comment := &entity.Comment{Content: "test", PostID: 1, UserID: 1}
		repo.On("CreateComment", mock.Anything, comment).Return(nil)

		err := uc.CreateComment(context.Background(), comment)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("ValidationError", func(t *testing.T) {
		repo := new(mocks.MockCommentRepository)
		uc := usecase.NewCommentUseCase(repo)

		testCases := []struct {
			name    string
			comment *entity.Comment
			errMsg  string
		}{
			{
				name:    "EmptyContent",
				comment: &entity.Comment{Content: "", PostID: 1, UserID: 1},
				errMsg:  "comment content cannot be empty",
			},
			{
				name:    "ZeroPostID",
				comment: &entity.Comment{Content: "test", PostID: 0, UserID: 1},
				errMsg:  "post ID cannot be empty",
			},
			{
				name:    "ZeroUserID",
				comment: &entity.Comment{Content: "test", PostID: 1, UserID: 0},
				errMsg:  "user ID cannot be empty",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := uc.CreateComment(context.Background(), tc.comment)
				assert.ErrorContains(t, err, tc.errMsg)
			})
		}
	})

	t.Run("RepositoryError", func(t *testing.T) {
		repo := new(mocks.MockCommentRepository)
		uc := usecase.NewCommentUseCase(repo)

		expectedErr := errors.New("database error")
		comment := &entity.Comment{Content: "test", PostID: 1, UserID: 1}
		repo.On("CreateComment", mock.Anything, comment).Return(expectedErr)

		err := uc.CreateComment(context.Background(), comment)
		assert.ErrorIs(t, err, expectedErr)
		repo.AssertExpectations(t)
	})
}

func TestCommentUseCase_GetCommentsByPostID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := new(mocks.MockCommentRepository)
		uc := usecase.NewCommentUseCase(repo)

		expected := []entity.Comment{
			{ID: 1, Content: "test1"},
			{ID: 2, Content: "test2"},
		}

		repo.On("GetCommentsByPostID", mock.Anything, 1).Return(expected, nil)

		comments, err := uc.GetCommentsByPostID(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, expected, comments)
		repo.AssertExpectations(t)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		repo := new(mocks.MockCommentRepository)
		uc := usecase.NewCommentUseCase(repo)

		expectedErr := errors.New("database error")
		repo.On("GetCommentsByPostID", mock.Anything, 1).Return([]entity.Comment{}, expectedErr)

		comments, err := uc.GetCommentsByPostID(context.Background(), 1)
		assert.ErrorIs(t, err, expectedErr)
		assert.Empty(t, comments)
		repo.AssertExpectations(t)
	})

	t.Run("InvalidPostID", func(t *testing.T) {
		repo := new(mocks.MockCommentRepository)
		uc := usecase.NewCommentUseCase(repo)

		_, err := uc.GetCommentsByPostID(context.Background(), 0)
		assert.ErrorContains(t, err, "invalid post ID")
	})
}

func TestCommentUseCase_DeleteComment(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := new(mocks.MockCommentRepository)
		uc := usecase.NewCommentUseCase(repo)

		repo.On("DeleteComment", mock.Anything, 1, 1).Return(nil)

		err := uc.DeleteComment(context.Background(), 1, 1)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		repo := new(mocks.MockCommentRepository)
		uc := usecase.NewCommentUseCase(repo)

		expectedErr := errors.New("database error")
		repo.On("DeleteComment", mock.Anything, 1, 1).Return(expectedErr)

		err := uc.DeleteComment(context.Background(), 1, 1)
		assert.ErrorIs(t, err, expectedErr)
		repo.AssertExpectations(t)
	})

	t.Run("InvalidIDs", func(t *testing.T) {
		repo := new(mocks.MockCommentRepository)
		uc := usecase.NewCommentUseCase(repo)

		testCases := []struct {
			name      string
			commentID int
			userID    int
			errMsg    string
		}{
			{"ZeroCommentID", 0, 1, "invalid comment ID"},
			{"ZeroUserID", 1, 0, "invalid user ID"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := uc.DeleteComment(context.Background(), tc.commentID, tc.userID)
				assert.ErrorContains(t, err, tc.errMsg)
			})
		}
	})
}
