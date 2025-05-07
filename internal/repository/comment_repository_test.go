// internal/repository/comment_repository_test.go
package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCommentRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateComment", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(mocks.MockCommentRepository)
			comment := &entity.Comment{
				Content: "Test comment",
				PostID:  1,
				UserID:  1,
			}

			mockRepo.On("CreateComment", ctx, comment).Return(nil)

			err := mockRepo.CreateComment(ctx, comment)
			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})

		t.Run("Error", func(t *testing.T) {
			mockRepo := new(mocks.MockCommentRepository)
			comment := &entity.Comment{
				Content: "Test comment",
				PostID:  1,
				UserID:  1,
			}
			expectedErr := errors.New("database error")

			mockRepo.On("CreateComment", ctx, comment).Return(expectedErr)

			err := mockRepo.CreateComment(ctx, comment)
			assert.ErrorIs(t, err, expectedErr)
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("GetCommentsByPostID", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(mocks.MockCommentRepository)
			postID := 1
			expectedComments := []entity.Comment{
				{
					ID:        1,
					Content:   "Comment 1",
					PostID:    postID,
					UserID:    1,
					Author:    "user1",
					CreatedAt: time.Now(),
				},
				{
					ID:        2,
					Content:   "Comment 2",
					PostID:    postID,
					UserID:    2,
					Author:    "user2",
					CreatedAt: time.Now(),
				},
			}

			mockRepo.On("GetCommentsByPostID", ctx, postID).Return(expectedComments, nil)

			comments, err := mockRepo.GetCommentsByPostID(ctx, postID)
			assert.NoError(t, err)
			assert.Equal(t, expectedComments, comments)
			mockRepo.AssertExpectations(t)
		})

		t.Run("EmptyResult", func(t *testing.T) {
			mockRepo := new(mocks.MockCommentRepository)
			postID := 1

			mockRepo.On("GetCommentsByPostID", ctx, postID).Return([]entity.Comment{}, nil)

			comments, err := mockRepo.GetCommentsByPostID(ctx, postID)
			assert.NoError(t, err)
			assert.Empty(t, comments)
			mockRepo.AssertExpectations(t)
		})

		t.Run("Error", func(t *testing.T) {
			mockRepo := new(mocks.MockCommentRepository)
			postID := 1
			expectedErr := errors.New("database error")

			mockRepo.On("GetCommentsByPostID", ctx, postID).Return(nil, expectedErr)

			comments, err := mockRepo.GetCommentsByPostID(ctx, postID)
			assert.ErrorIs(t, err, expectedErr)
			assert.Nil(t, comments)
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("DeleteComment", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(mocks.MockCommentRepository)
			commentID := 1
			userID := 1

			mockRepo.On("DeleteComment", ctx, commentID, userID).Return(nil)

			err := mockRepo.DeleteComment(ctx, commentID, userID)
			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})

		t.Run("NotFound", func(t *testing.T) {
			mockRepo := new(mocks.MockCommentRepository)
			commentID := 1
			userID := 1
			expectedErr := sql.ErrNoRows

			mockRepo.On("DeleteComment", ctx, commentID, userID).Return(expectedErr)

			err := mockRepo.DeleteComment(ctx, commentID, userID)
			assert.ErrorIs(t, err, expectedErr)
			mockRepo.AssertExpectations(t)
		})

		t.Run("DatabaseError", func(t *testing.T) {
			mockRepo := new(mocks.MockCommentRepository)
			commentID := 1
			userID := 1
			expectedErr := errors.New("database error")

			mockRepo.On("DeleteComment", ctx, commentID, userID).Return(expectedErr)

			err := mockRepo.DeleteComment(ctx, commentID, userID)
			assert.ErrorIs(t, err, expectedErr)
			mockRepo.AssertExpectations(t)
		})
	})
}
