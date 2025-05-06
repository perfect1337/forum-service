// internal/usecase/post_test.go
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

func TestPostUseCase_CreatePost(t *testing.T) {
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	uc := usecase.NewPostUseCase(mockPostRepo, mockUserRepo)

	t.Run("Success", func(t *testing.T) {
		post := &entity.Post{
			Title:   "Test Post",
			Content: "Test Content",
			UserID:  1,
		}
		mockPostRepo.On("CreatePost", mock.Anything, mock.AnythingOfType("*entity.Post")).Return(nil)

		err := uc.CreatePost(context.Background(), post)
		assert.NoError(t, err)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		post := &entity.Post{
			Title:   "Test Post",
			Content: "Test Content",
			UserID:  1,
		}
		expectedErr := errors.New("database error")

		// Используем точное сравнение аргументов
		mockPostRepo.On("CreatePost", mock.Anything, post).Return(expectedErr)

		err := uc.CreatePost(context.Background(), post)

		assert.ErrorIs(t, err, expectedErr)
		mockPostRepo.AssertExpectations(t)
	})

	// Тесты валидации
	t.Run("ValidationError_NilPost", func(t *testing.T) {
		err := uc.CreatePost(context.Background(), nil)
		assert.ErrorContains(t, err, "post cannot be nil")
	})

	t.Run("ValidationError_EmptyTitle", func(t *testing.T) {
		post := &entity.Post{
			Title:   "",
			Content: "Content",
			UserID:  1,
		}
		err := uc.CreatePost(context.Background(), post)
		assert.ErrorContains(t, err, "title cannot be empty")
	})

	t.Run("ValidationError_EmptyContent", func(t *testing.T) {
		post := &entity.Post{
			Title:   "Title",
			Content: "",
			UserID:  1,
		}
		err := uc.CreatePost(context.Background(), post)
		assert.ErrorContains(t, err, "content cannot be empty")
	})

	t.Run("ValidationError_EmptyUserID", func(t *testing.T) {
		post := &entity.Post{
			Title:   "Title",
			Content: "Content",
			UserID:  0,
		}
		err := uc.CreatePost(context.Background(), post)
		assert.ErrorContains(t, err, "user ID cannot be empty")
	})
}
func TestPostUseCase_DeletePost(t *testing.T) {
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	uc := usecase.NewPostUseCase(mockPostRepo, mockUserRepo)

	t.Run("Unauthorized", func(t *testing.T) {
		post := &entity.Post{ID: 1, UserID: 2}
		user := &entity.User{ID: 1, Role: "user"}

		mockPostRepo.On("GetPostByID", mock.Anything, 1).Return(post, nil)
		mockUserRepo.On("GetUserByID", mock.Anything, 1).Return(user, nil)

		err := uc.DeletePost(context.Background(), 1, 1)
		assert.ErrorContains(t, err, "unauthorized")
		mockPostRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("SuccessAdmin", func(t *testing.T) {
		post := &entity.Post{ID: 1, UserID: 2}
		admin := &entity.User{ID: 1, Role: "admin"}

		mockPostRepo.On("GetPostByID", mock.Anything, 1).Return(post, nil)
		mockUserRepo.On("GetUserByID", mock.Anything, 1).Return(admin, nil)
		mockPostRepo.On("DeletePost", mock.Anything, 1).Return(nil)

		err := uc.DeletePost(context.Background(), 1, 1)
		assert.NoError(t, err)
		mockPostRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}
