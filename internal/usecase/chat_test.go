package usecase_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Integration Tests
func TestChatUseCase_Integration(t *testing.T) {
	// Setup test database
	repo, err := repository.setupTestDB()
	require.NoError(t, err, "Failed to setup test database")

	authUC := &mockAuthUC{}
	uc := usecase.NewChatUseCase(repo, authUC)

	ctx := context.Background()

	t.Run("SendAndGetMessages", func(t *testing.T) {
		// Send test message
		msg := &entity.ChatMessage{
			UserID:    1,
			Author:    "testuser",
			Text:      "Hello world",
			CreatedAt: time.Now(),
		}

		err := uc.SendMessage(ctx, msg)
		require.NoError(t, err)
		assert.NotZero(t, msg.ID)

		// Get messages
		messages, err := uc.GetMessages(ctx, 10)
		require.NoError(t, err)
		require.Len(t, messages, 1)
		assert.Equal(t, "Hello world", messages[0].Text)
	})

	t.Run("CleanupOldMessages", func(t *testing.T) {
		// Insert old message
		oldMsg := &entity.ChatMessage{
			UserID:    1,
			Author:    "testuser",
			Text:      "Old message",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		}
		err := repo.SaveChatMessage(ctx, oldMsg)
		require.NoError(t, err)

		// Get messages should cleanup old ones
		messages, err := uc.GetMessages(ctx, 10)
		require.NoError(t, err)
		assert.Len(t, messages, 1) // Only the new message should remain
	})
}

// WebSocket tests need a real WebSocket connection
func TestChatUseCase_WebSocketHandler(t *testing.T) {
	// Setup test database
	repo, err := repository.setupTestDB()
	require.NoError(t, err, "Failed to setup test database")

	authUC := &mockAuthUC{}
	uc := usecase.NewChatUseCase(repo, authUC)

	// Create test server
	server := httptest.NewServer(uc.HandleWebSocket)
	defer server.Close()

	// Create WebSocket connection
	wsURL := "ws" + server.URL[4:] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Skip("Skipping WebSocket test - could not connect to test server")
	}
	defer conn.Close()

	t.Run("HandleValidMessage", func(t *testing.T) {
		authUC.On("ParseToken", "valid-token").Return(int64(1), "testuser", nil)

		// Send test message
		testMsg := `{"text": "ws test", "token": "valid-token"}`
		err := conn.WriteMessage(websocket.TextMessage, []byte(testMsg))
		require.NoError(t, err)

		// Read response
		_, msg, err := conn.ReadMessage()
		require.NoError(t, err)
		assert.Contains(t, string(msg), "ws test")

		// Verify message was saved
		messages, err := repo.GetChatMessages(context.Background(), 10)
		require.NoError(t, err)
		assert.Contains(t, messages[len(messages)-1].Text, "ws test")
	})
}

// Mock Tests
type MockChatRepository struct {
	mock.Mock
}

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

func TestChatUseCase_SendMessage(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockChatRepository)
		authUC := new(mockAuthUC)
		uc := usecase.NewChatUseCase(mockRepo, authUC)

		msg := &entity.ChatMessage{Text: "test"}
		mockRepo.On("SaveChatMessage", mock.Anything, msg).Return(nil)

		err := uc.SendMessage(context.Background(), msg)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("RepositoryError", func(t *testing.T) {
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

func TestChatUseCase_GetMessages(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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

	t.Run("CleanupError", func(t *testing.T) {
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

	t.Run("GetMessagesError", func(t *testing.T) {
		mockRepo := new(MockChatRepository)
		authUC := new(mockAuthUC)
		uc := usecase.NewChatUseCase(mockRepo, authUC)

		getErr := errors.New("get messages error")
		mockRepo.On("DeleteOldChatMessages", mock.Anything, 30*time.Minute).Return(nil)
		mockRepo.On("GetChatMessages", mock.Anything, 100).Return(nil, getErr)

		_, err := uc.GetMessages(context.Background(), 100)
		assert.Error(t, err)
		assert.Equal(t, getErr, err)
		mockRepo.AssertExpectations(t)
	})
}
