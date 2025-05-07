package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("GetUserByID", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			userID := 1
			expectedUser := &entity.User{
				ID:       userID,
				Username: "testuser",
				Email:    "test@example.com",
				Role:     "user",
			}

			mockRepo.On("GetUserByID", ctx, userID).Return(expectedUser, nil)

			user, err := mockRepo.GetUserByID(ctx, userID)
			assert.NoError(t, err)
			assert.Equal(t, expectedUser, user)
			mockRepo.AssertExpectations(t)
		})

		t.Run("NotFound", func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			userID := 999

			mockRepo.On("GetUserByID", ctx, userID).Return(nil, errors.New("user not found"))

			user, err := mockRepo.GetUserByID(ctx, userID)
			assert.Error(t, err)
			assert.Nil(t, user)
			mockRepo.AssertExpectations(t)
		})

		t.Run("DatabaseError", func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			userID := 1
			expectedErr := errors.New("database error")

			mockRepo.On("GetUserByID", ctx, userID).Return(nil, expectedErr)

			user, err := mockRepo.GetUserByID(ctx, userID)
			assert.ErrorIs(t, err, expectedErr)
			assert.Nil(t, user)
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("GetUsersByIDs", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			userIDs := []int{1, 2}
			expectedUsers := map[int]*entity.User{
				1: {
					ID:       1,
					Username: "user1",
					Email:    "user1@example.com",
					Role:     "user",
				},
				2: {
					ID:       2,
					Username: "user2",
					Email:    "user2@example.com",
					Role:     "admin",
				},
			}

			mockRepo.On("GetUsersByIDs", ctx, userIDs).Return(expectedUsers, nil)

			users, err := mockRepo.GetUsersByIDs(ctx, userIDs)
			assert.NoError(t, err)
			assert.Equal(t, expectedUsers, users)
			mockRepo.AssertExpectations(t)
		})

		t.Run("EmptyResult", func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			userIDs := []int{999, 1000}

			mockRepo.On("GetUsersByIDs", ctx, userIDs).Return(map[int]*entity.User{}, nil)

			users, err := mockRepo.GetUsersByIDs(ctx, userIDs)
			assert.NoError(t, err)
			assert.Empty(t, users)
			mockRepo.AssertExpectations(t)
		})

		t.Run("DatabaseError", func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			userIDs := []int{1, 2}
			expectedErr := errors.New("database error")

			mockRepo.On("GetUsersByIDs", ctx, userIDs).Return(nil, expectedErr)

			users, err := mockRepo.GetUsersByIDs(ctx, userIDs)
			assert.ErrorIs(t, err, expectedErr)
			assert.Nil(t, users)
			mockRepo.AssertExpectations(t)
		})
	})
}
