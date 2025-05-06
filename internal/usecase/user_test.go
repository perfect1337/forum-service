package usecase_test

import (
	"context"
	"testing"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/mocks"
	"github.com/perfect1337/forum-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserUseCase_GetUserByID(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	uc := usecase.NewUserUseCase(repo)

	t.Run("Found", func(t *testing.T) {
		expected := &entity.User{ID: 1}
		repo.On("GetUserByID", mock.Anything, 1).Return(expected, nil)

		user, err := uc.GetUserByID(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, expected, user)
	})
}
