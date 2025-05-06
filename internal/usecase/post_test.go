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
		post := &entity.Post{Title: "Test"}
		mockPostRepo.On("CreatePost", mock.Anything, post).Return(nil)

		err := uc.CreatePost(context.Background(), post)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		post := &entity.Post{Title: "Test"}
		mockPostRepo.On("CreatePost", mock.Anything, post).Return(errors.New("db error"))

		err := uc.CreatePost(context.Background(), post)
		assert.ErrorContains(t, err, "db error")
	})
}

func TestPostUseCase_DeletePost(t *testing.T) {
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	uc := usecase.NewPostUseCase(mockPostRepo, mockUserRepo)

	t.Run("Unauthorized", func(t *testing.T) {
		mockPostRepo.On("GetPostByID", mock.Anything, 1).Return(&entity.Post{UserID: 2}, nil)
		mockUserRepo.On("GetUserByID", mock.Anything, 1).Return(&entity.User{Role: "user"}, nil)

		err := uc.DeletePost(context.Background(), 1, 1)
		assert.ErrorContains(t, err, "unauthorized")
	})
}
