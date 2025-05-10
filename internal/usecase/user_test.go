// internal/usecase/user_test.go
package usecase_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/repository"
	"github.com/perfect1337/forum-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// InMemoryUserRepository реализация для интеграционных тестов
type InMemoryUserRepository struct {
	users map[int]*entity.User
	mu    sync.RWMutex
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[int]*entity.User),
	}
}

func (r *InMemoryUserRepository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *InMemoryUserRepository) GetUsersByIDs(ctx context.Context, ids []int) (map[int]*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[int]*entity.User)
	for _, id := range ids {
		if user, ok := r.users[id]; ok {
			result[id] = user
		}
	}
	return result, nil
}

func (r *InMemoryUserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

// MockUserRepository для модульных тестов
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUsersByIDs(ctx context.Context, ids []int) (map[int]*entity.User, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(map[int]*entity.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func TestUserUseCase_GetUserByID(t *testing.T) {
	repo := new(MockUserRepository)
	uc := usecase.NewUserUseCase(repo)

	t.Run("Found", func(t *testing.T) {
		expected := &entity.User{ID: 1, Username: "test"}
		repo.On("GetUserByID", mock.Anything, 1).Return(expected, nil)

		user, err := uc.GetUserByID(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, expected, user)
	})

	t.Run("Not Found", func(t *testing.T) {
		repo.On("GetUserByID", mock.Anything, 2).Return((*entity.User)(nil), repository.ErrNotFound)

		_, err := uc.GetUserByID(context.Background(), 2)
		assert.Error(t, err)
		assert.Equal(t, repository.ErrNotFound, err)
	})
}

func TestUserUseCase_GetUsersByIDs(t *testing.T) {
	t.Run("Integration - Successful", func(t *testing.T) {
		// Создаем in-memory репозиторий
		repo := NewInMemoryUserRepository()
		uc := usecase.NewUserUseCase(repo)

		// Добавляем тестовых пользователей
		testUsers := []*entity.User{
			{ID: 1, Username: "user1"},
			{ID: 2, Username: "user2"},
			{ID: 3, Username: "user3"},
		}

		for _, user := range testUsers {
			err := repo.CreateUser(context.Background(), user)
			assert.NoError(t, err)
		}

		// Тестируем получение нескольких пользователей
		users, err := uc.GetUsersByIDs(context.Background(), []int{1, 2})
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "user1", users[1].Username)
		assert.Equal(t, "user2", users[2].Username)
	})

	t.Run("Unit - Successful", func(t *testing.T) {
		repo := new(MockUserRepository)
		uc := usecase.NewUserUseCase(repo)

		expected := map[int]*entity.User{
			1: {ID: 1, Username: "user1"},
			2: {ID: 2, Username: "user2"},
		}
		repo.On("GetUsersByIDs", mock.Anything, []int{1, 2}).Return(expected, nil)

		users, err := uc.GetUsersByIDs(context.Background(), []int{1, 2})
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "user1", users[1].Username)
		assert.Equal(t, "user2", users[2].Username)
	})

	t.Run("Empty IDs", func(t *testing.T) {
		repo := new(MockUserRepository)
		uc := usecase.NewUserUseCase(repo)

		repo.On("GetUsersByIDs", mock.Anything, []int{}).Return(map[int]*entity.User{}, nil)

		users, err := uc.GetUsersByIDs(context.Background(), []int{})
		assert.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("Partial Found", func(t *testing.T) {
		repo := new(MockUserRepository)
		uc := usecase.NewUserUseCase(repo)

		expected := map[int]*entity.User{
			1: {ID: 1, Username: "user1"},
		}
		repo.On("GetUsersByIDs", mock.Anything, []int{1, 999}).Return(expected, nil)

		users, err := uc.GetUsersByIDs(context.Background(), []int{1, 999})
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "user1", users[1].Username)
	})
}
