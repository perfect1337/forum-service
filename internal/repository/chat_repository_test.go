package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestChatRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateChatMessage", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(mocks.MockChatRepository)
			message := &entity.ChatMessage{
				UserID: 1,
				Author: "user1",
				Text:   "Hello, World!",
			}

			mockRepo.On("CreateChatMessage", ctx, message).Return(nil)

			err := mockRepo.CreateChatMessage(ctx, message)
			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})

		t.Run("Error", func(t *testing.T) {
			mockRepo := new(mocks.MockChatRepository)
			message := &entity.ChatMessage{
				UserID: 1,
				Author: "user1",
				Text:   "Hello, World!",
			}
			expectedErr := errors.New("database error")

			mockRepo.On("CreateChatMessage", ctx, message).Return(expectedErr)

			err := mockRepo.CreateChatMessage(ctx, message)
			assert.ErrorIs(t, err, expectedErr)
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("GetChatMessages", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(mocks.MockChatRepository)
			limit := 10
			expectedMessages := []entity.ChatMessage{
				{ID: 1, UserID: 1, Author: "user1", Text: "Hello, World!", CreatedAt: time.Now()},
				{ID: 2, UserID: 2, Author: "user2", Text: "Hi there!", CreatedAt: time.Now()},
			}

			mockRepo.On("GetChatMessages", ctx, limit).Return(expectedMessages, nil)

			messages, err := mockRepo.GetChatMessages(ctx, limit)
			assert.NoError(t, err)
			assert.Equal(t, expectedMessages, messages)
			mockRepo.AssertExpectations(t)
		})

		t.Run("Error", func(t *testing.T) {
			mockRepo := new(mocks.MockChatRepository)
			limit := 10
			expectedErr := errors.New("database error")

			mockRepo.On("GetChatMessages", ctx, limit).Return([]entity.ChatMessage{}, expectedErr)

			messages, err := mockRepo.GetChatMessages(ctx, limit)
			assert.ErrorIs(t, err, expectedErr)
			assert.Empty(t, messages)
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("SaveChatMessage", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(mocks.MockChatRepository)
			message := &entity.ChatMessage{
				UserID: 1,
				Author: "user1",
				Text:   "Hello, World!",
			}

			mockRepo.On("SaveChatMessage", ctx, message).Return(nil)

			err := mockRepo.SaveChatMessage(ctx, message)
			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})

		t.Run("Error", func(t *testing.T) {
			mockRepo := new(mocks.MockChatRepository)
			message := &entity.ChatMessage{
				UserID: 1,
				Author: "user1",
				Text:   "Hello, World!",
			}
			expectedErr := errors.New("database error")

			mockRepo.On("SaveChatMessage", ctx, message).Return(expectedErr)

			err := mockRepo.SaveChatMessage(ctx, message)
			assert.ErrorIs(t, err, expectedErr)
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("DeleteOldChatMessages", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(mocks.MockChatRepository)
			olderThan := 24 * time.Hour

			mockRepo.On("DeleteOldChatMessages", ctx, olderThan).Return(nil)

			err := mockRepo.DeleteOldChatMessages(ctx, olderThan)
			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})

		t.Run("Error", func(t *testing.T) {
			mockRepo := new(mocks.MockChatRepository)
			olderThan := 24 * time.Hour
			expectedErr := errors.New("database error")

			mockRepo.On("DeleteOldChatMessages", ctx, olderThan).Return(expectedErr)

			err := mockRepo.DeleteOldChatMessages(ctx, olderThan)
			assert.ErrorIs(t, err, expectedErr)
			mockRepo.AssertExpectations(t)
		})
	})
}
