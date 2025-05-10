// internal/usecase/chat_test.go
package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChatRepository мокает репозиторий чата
type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) SaveChatMessage(ctx context.Context, msg *entity.ChatMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockChatRepository) GetChatMessages(ctx context.Context, limit int) ([]entity.ChatMessage, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]entity.ChatMessage), args.Error(1)
}

func (m *MockChatRepository) DeleteOldChatMessages(ctx context.Context, olderThan time.Duration) error {
	args := m.Called(ctx, olderThan)
	return args.Error(0)
}

// mockAuthUC мокает AuthUseCase
type mockAuthUC struct {
	mock.Mock
}

func (m *mockAuthUC) SecretKey() []byte {
	return []byte("test-secret")
}

func (m *mockAuthUC) GenerateToken(userID int, username string) (string, error) {
	args := m.Called(userID, username)
	return args.String(0), args.Error(1)
}

func (m *mockAuthUC) ParseToken(tokenString string) (int64, string, error) {
	args := m.Called(tokenString)
	return args.Get(0).(int64), args.String(1), args.Error(2)
}

// TestChatUseCase_SendMessage тестирует отправку сообщений
func TestChatUseCase_SendMessage(t *testing.T) {
	t.Run("Успешная отправка", func(t *testing.T) {
		mockRepo := new(MockChatRepository)
		authUC := new(mockAuthUC)
		uc := usecase.NewChatUseCase(mockRepo, authUC)

		msg := &entity.ChatMessage{Text: "test"}
		mockRepo.On("SaveChatMessage", mock.Anything, msg).Return(nil)

		err := uc.SendMessage(context.Background(), msg)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Ошибка репозитория", func(t *testing.T) {
		mockRepo := new(MockChatRepository)
		authUC := new(mockAuthUC)
		uc := usecase.NewChatUseCase(mockRepo, authUC)

		msg := &entity.ChatMessage{Text: "test"}
		repoErr := errors.New("repository error")
		mockRepo.On("SaveChatMessage", mock.Anything, msg).Return(repoErr)

		err := uc.SendMessage(context.Background(), msg)
		assert.Error(t, err)
		assert.Equal(t, repoErr, err)
		mockRepo.AssertExpectations(t)
	})
}

// TestChatUseCase_GetMessages тестирует получение сообщений
func TestChatUseCase_GetMessages(t *testing.T) {
	t.Run("Успешное получение", func(t *testing.T) {
		mockRepo := new(MockChatRepository)
		authUC := new(mockAuthUC)
		uc := usecase.NewChatUseCase(mockRepo, authUC)

		testMessages := []entity.ChatMessage{
			{Text: "message 1"},
			{Text: "message 2"},
		}

		mockRepo.On("DeleteOldChatMessages", mock.Anything, 30*time.Minute).Return(nil)
		mockRepo.On("GetChatMessages", mock.Anything, 100).Return(testMessages, nil)

		messages, err := uc.GetMessages(context.Background(), 100)
		assert.NoError(t, err)
		assert.Len(t, messages, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Ошибка очистки", func(t *testing.T) {
		mockRepo := new(MockChatRepository)
		authUC := new(mockAuthUC)
		uc := usecase.NewChatUseCase(mockRepo, authUC)

		cleanupErr := errors.New("cleanup error")
		mockRepo.On("DeleteOldChatMessages", mock.Anything, 30*time.Minute).Return(cleanupErr)

		_, err := uc.GetMessages(context.Background(), 100)
		assert.Error(t, err)
		assert.Equal(t, cleanupErr, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Ошибка получения", func(t *testing.T) {
		mockRepo := new(MockChatRepository)
		authUC := new(mockAuthUC)
		uc := usecase.NewChatUseCase(mockRepo, authUC)

		getErr := errors.New("get messages error")
		mockRepo.On("DeleteOldChatMessages", mock.Anything, 30*time.Minute).Return(nil)
		mockRepo.On("GetChatMessages", mock.Anything, 100).Return([]entity.ChatMessage{}, getErr)

		_, err := uc.GetMessages(context.Background(), 100)
		assert.Error(t, err)
		assert.Equal(t, getErr, err)
		mockRepo.AssertExpectations(t)
	})
}

// TestAuthUseCase тестирует методы аутентификации
func TestAuthUseCase(t *testing.T) {
	t.Run("Генерация токена", func(t *testing.T) {
		authUC := new(mockAuthUC)
		authUC.On("GenerateToken", 1, "user1").Return("test-token", nil)

		token, err := authUC.GenerateToken(1, "user1")
		assert.NoError(t, err)
		assert.Equal(t, "test-token", token)
	})

	t.Run("Парсинг токена", func(t *testing.T) {
		authUC := new(mockAuthUC)
		authUC.On("ParseToken", "valid-token").Return(int64(1), "user1", nil)

		userID, username, err := authUC.ParseToken("valid-token")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), userID)
		assert.Equal(t, "user1", username)
	})
}
