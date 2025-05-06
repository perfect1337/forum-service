// internal/usecase/chat_test.go
package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/mocks"
	"github.com/perfect1337/forum-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChatUseCase_SendMessage(t *testing.T) {
	mockChatRepo := new(mocks.MockChatRepository)
	mockAuthUC := new(mocks.MockAuthUseCase)
	uc := usecase.NewChatUseCase(mockChatRepo, mockAuthUC)

	t.Run("Success", func(t *testing.T) {
		msg := &entity.ChatMessage{Text: "test"}
		mockChatRepo.On("SaveChatMessage", mock.Anything, msg).Return(nil)

		err := uc.SendMessage(context.Background(), msg)
		assert.NoError(t, err)
		mockChatRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		msg := &entity.ChatMessage{Text: "test"}
		mockChatRepo.On("SaveChatMessage", mock.Anything, msg).Return(assert.AnError)

		err := uc.SendMessage(context.Background(), msg)
		assert.Error(t, err)
		mockChatRepo.AssertExpectations(t)
	})
}

func TestChatUseCase_GetMessages(t *testing.T) {
	mockChatRepo := new(mocks.MockChatRepository)
	uc := usecase.NewChatUseCase(mockChatRepo, nil)

	t.Run("Success", func(t *testing.T) {
		expected := []entity.ChatMessage{{Text: "test"}}
		mockChatRepo.On("DeleteOldChatMessages", mock.Anything, 30*time.Minute).Return(nil)
		mockChatRepo.On("GetChatMessages", mock.Anything, 100).Return(expected, nil)

		result, err := uc.GetMessages(context.Background(), 100)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockChatRepo.AssertExpectations(t)
	})
}
