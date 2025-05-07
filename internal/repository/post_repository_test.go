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

func TestPostRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("CreatePost", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(mocks.MockPostRepository)
			post := &entity.Post{
				Title:   "Test Post",
				Content: "Test Content",
				UserID:  1,
			}

			mockRepo.On("CreatePost", ctx, post).Return(nil)

			err := mockRepo.CreatePost(ctx, post)
			assert.NoError(t, err)
			assert.NotZero(t, post.ID) // Assuming ID is set by CreatePost
			mockRepo.AssertExpectations(t)
		})

		t.Run("Error", func(t *testing.T) {
			mockRepo := new(mocks.MockPostRepository)
			post := &entity.Post{
				Title:   "Test Post",
				Content: "Test Content",
				UserID:  1,
			}
			expectedErr := errors.New("database error")

			mockRepo.On("CreatePost", ctx, post).Return(expectedErr)

			err := mockRepo.CreatePost(ctx, post)
			assert.ErrorIs(t, err, expectedErr)
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("GetAllPosts", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(mocks.MockPostRepository)
			post := &entity.Post{
				Title:   "Test Post",
				Content: "Test Content",
				UserID:  1,
			}

			mockRepo.On("CreatePost", ctx, post).Return(nil)

			err := mockRepo.CreatePost(ctx, post)
			assert.NoError(t, err)
			assert.NotZero(t, post.ID) // This assertion is failing
			mockRepo.AssertExpectations(t)
		})
		t.Run("EmptyResult", func(t *testing.T) {
			mockRepo := new(mocks.MockPostRepository)

			mockRepo.On("GetAllPosts", ctx).Return([]*entity.Post{}, nil)

			posts, err := mockRepo.GetAllPosts(ctx)
			assert.NoError(t, err)
			assert.Empty(t, posts)
			mockRepo.AssertExpectations(t)
		})

		t.Run("Error", func(t *testing.T) {
			mockRepo := new(mocks.MockPostRepository)
			expectedErr := errors.New("database error")

			mockRepo.On("GetAllPosts", ctx).Return(nil, expectedErr)

			posts, err := mockRepo.GetAllPosts(ctx)
			assert.ErrorIs(t, err, expectedErr)
			assert.Nil(t, posts)
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("GetPostByID", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(mocks.MockPostRepository)
			postID := 1
			expectedPost := &entity.Post{
				ID:        postID,
				Title:     "Test Post",
				Content:   "Test Content",
				UserID:    1,
				Author:    "user1",
				CreatedAt: time.Now(),
			}

			mockRepo.On("GetPostByID", ctx, postID).Return(expectedPost, nil)

			post, err := mockRepo.GetPostByID(ctx, postID)
			assert.NoError(t, err)
			assert.Equal(t, expectedPost, post)
			mockRepo.AssertExpectations(t)
		})

		t.Run("NotFound", func(t *testing.T) {
			mockRepo := new(mocks.MockPostRepository)
			postID := 999
			expectedErr := sql.ErrNoRows

			mockRepo.On("GetPostByID", ctx, postID).Return(nil, expectedErr)

			post, err := mockRepo.GetPostByID(ctx, postID)
			assert.ErrorIs(t, err, expectedErr)
			assert.Nil(t, post)
			mockRepo.AssertExpectations(t)
		})

		t.Run("Error", func(t *testing.T) {
			mockRepo := new(mocks.MockPostRepository)
			postID := 1
			expectedErr := errors.New("database error")

			mockRepo.On("GetPostByID", ctx, postID).Return(nil, expectedErr)

			post, err := mockRepo.GetPostByID(ctx, postID)
			assert.ErrorIs(t, err, expectedErr)
			assert.Nil(t, post)
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("DeletePost", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(mocks.MockPostRepository)
			postID := 1

			mockRepo.On("DeletePost", ctx, postID).Return(nil)

			err := mockRepo.DeletePost(ctx, postID)
			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})

		t.Run("NotFound", func(t *testing.T) {
			mockRepo := new(mocks.MockPostRepository)
			postID := 999
			expectedErr := sql.ErrNoRows

			mockRepo.On("DeletePost", ctx, postID).Return(expectedErr)

			err := mockRepo.DeletePost(ctx, postID)
			assert.ErrorIs(t, err, expectedErr)
			mockRepo.AssertExpectations(t)
		})

		t.Run("Error", func(t *testing.T) {
			mockRepo := new(mocks.MockPostRepository)
			postID := 1
			expectedErr := errors.New("database error")

			mockRepo.On("DeletePost", ctx, postID).Return(expectedErr)

			err := mockRepo.DeletePost(ctx, postID)
			assert.ErrorIs(t, err, expectedErr)
			mockRepo.AssertExpectations(t)
		})
	})
}
