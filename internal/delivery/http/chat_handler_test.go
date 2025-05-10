// internal/delivery/http/chat_handler_test.go
package delivery_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	delivery "github.com/perfect1337/forum-service/internal/delivery/http"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/usecase"

	"github.com/stretchr/testify/assert"
)

type MockChatUseCase struct {
	HandleWebSocketFunc func(conn usecase.WebSocketConnection)
	SendMessageFunc     func(ctx context.Context, message *entity.ChatMessage) error
	GetMessagesFunc     func(ctx context.Context, limit int) ([]entity.ChatMessage, error)
}

func (m *MockChatUseCase) HandleWebSocket(conn usecase.WebSocketConnection) {
	if m.HandleWebSocketFunc != nil {
		m.HandleWebSocketFunc(conn)
	}
}

func (m *MockChatUseCase) SendMessage(ctx context.Context, message *entity.ChatMessage) error {
	if m.SendMessageFunc != nil {
		return m.SendMessageFunc(ctx, message)
	}
	return nil
}

func (m *MockChatUseCase) GetMessages(ctx context.Context, limit int) ([]entity.ChatMessage, error) {
	if m.GetMessagesFunc != nil {
		return m.GetMessagesFunc(ctx, limit)
	}
	return nil, nil
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func TestChatHandler_HandleWebSocket(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		mockUC := &MockChatUseCase{
			HandleWebSocketFunc: func(conn usecase.WebSocketConnection) {
				// Тестовая логика
			},
		}

		handler := delivery.NewChatHandler(mockUC)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := gin.CreateTestContextOnly(w, &gin.Engine{})
			ctx.Request = r
			handler.HandleWebSocket(ctx)
		}))
		defer server.Close()

		_, _, err := websocket.DefaultDialer.Dial("ws"+server.URL[4:], nil)
		assert.NoError(t, err)
	})

	t.Run("failed upgrade", func(t *testing.T) {
		mockUC := &MockChatUseCase{}
		handler := delivery.NewChatHandler(mockUC)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/ws", nil)
		c.Request.Header = nil // Break the request

		handler.HandleWebSocket(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
func TestChatHandler_SendMessage(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(c *gin.Context)
		setupMock      func() *MockChatUseCase
		requestBody    string
		expectedStatus int
	}{
		{
			name: "successful message send",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", 123)
				c.Set("username", "testuser")
			},
			setupMock: func() *MockChatUseCase {
				return &MockChatUseCase{
					SendMessageFunc: func(ctx context.Context, message *entity.ChatMessage) error {
						return nil
					},
				}
			},
			requestBody:    `{"text": "Hello"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing user_id",
			setupContext: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			setupMock:      func() *MockChatUseCase { return &MockChatUseCase{} },
			requestBody:    `{"text": "Hello"}`,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "database error",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", 123)
				c.Set("username", "testuser")
			},
			setupMock: func() *MockChatUseCase {
				return &MockChatUseCase{
					SendMessageFunc: func(ctx context.Context, message *entity.ChatMessage) error {
						return errors.New("db error")
					},
				}
			},
			requestBody:    `{"text": "Hello"}`,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := tt.setupMock()
			handler := delivery.NewChatHandler(mockUC)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/messages", nil)
			if tt.requestBody != "" {
				c.Request.Body = nopCloser{strings.NewReader(tt.requestBody)}
				c.Request.Header.Set("Content-Type", "application/json")
			}

			tt.setupContext(c)
			handler.SendMessage(c)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestChatHandler_GetMessages(t *testing.T) {
	t.Run("successful get messages", func(t *testing.T) {
		mockMessages := []entity.ChatMessage{
			{ID: 1, UserID: 1, Author: "user1", Text: "Hello"},
		}

		mockUC := &MockChatUseCase{
			GetMessagesFunc: func(ctx context.Context, limit int) ([]entity.ChatMessage, error) {
				return mockMessages, nil
			},
		}

		handler := delivery.NewChatHandler(mockUC)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/messages", nil)

		handler.GetMessages(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("database error", func(t *testing.T) {
		mockUC := &MockChatUseCase{
			GetMessagesFunc: func(ctx context.Context, limit int) ([]entity.ChatMessage, error) {
				return nil, errors.New("db error")
			},
		}

		handler := delivery.NewChatHandler(mockUC)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/messages", nil)

		handler.GetMessages(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNewChatHandler(t *testing.T) {
	mockUC := &MockChatUseCase{}
	handler := delivery.NewChatHandler(mockUC)
	assert.NotNil(t, handler)
}
