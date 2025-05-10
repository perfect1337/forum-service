package usecase_test

import (
	"context"
	"testing"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUsersByIDs(ctx context.Context, ids []int) (map[int]*entity.User, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(map[int]*entity.User), args.Error(1)
}

func TestUserUseCase_GetUserByID(t *testing.T) {
	repo := new(MockUserRepository)
	uc := usecase.NewUserUseCase(repo)

	t.Run("Found", func(t *testing.T) {
		expected := &entity.User{ID: 1}
		repo.On("GetUserByID", mock.Anything, 1).Return(expected, nil)

		user, err := uc.GetUserByID(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, expected, user)
	})
}
