package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockCommentRepository struct {
	mock.Mock
}

func (m *MockCommentRepository) CreateComment(ctx context.Context, comment *entity.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentRepository) GetCommentsByPostID(ctx context.Context, postID int) ([]entity.Comment, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Comment), args.Error(1)
}

func (m *MockCommentRepository) DeleteComment(ctx context.Context, commentID, userID int) error {
	args := m.Called(ctx, commentID, userID)
	return args.Error(0)
}

func TestCommentUseCase_CreateComment(t *testing.T) {
	tests := []struct {
		name        string
		comment     *entity.Comment
		mockSetup   func(*MockCommentRepository)
		expectedErr string
	}{
		{
			name: "Success",
			comment: &entity.Comment{
				Content: "Test content",
				PostID:  1,
				UserID:  1,
			},
			mockSetup: func(m *MockCommentRepository) {
				m.On("CreateComment", mock.Anything, mock.AnythingOfType("*entity.Comment")).Return(nil)
			},
		},
		{
			name: "EmptyContent",
			comment: &entity.Comment{
				Content: "",
				PostID:  1,
				UserID:  1,
			},
			mockSetup:   func(m *MockCommentRepository) {},
			expectedErr: "comment content cannot be empty",
		},
		{
			name: "EmptyPostID",
			comment: &entity.Comment{
				Content: "Test content",
				PostID:  0,
				UserID:  1,
			},
			mockSetup:   func(m *MockCommentRepository) {},
			expectedErr: "post ID cannot be empty",
		},
		{
			name: "EmptyUserID",
			comment: &entity.Comment{
				Content: "Test content",
				PostID:  1,
				UserID:  0,
			},
			mockSetup:   func(m *MockCommentRepository) {},
			expectedErr: "user ID cannot be empty",
		},
		{
			name: "RepositoryError",
			comment: &entity.Comment{
				Content: "Test content",
				PostID:  1,
				UserID:  1,
			},
			mockSetup: func(m *MockCommentRepository) {
				m.On("CreateComment", mock.Anything, mock.AnythingOfType("*entity.Comment")).
					Return(errors.New("database error"))
			},
			expectedErr: "database error",
		},
		{
			name:        "NilComment",
			comment:     nil,
			mockSetup:   func(m *MockCommentRepository) {},
			expectedErr: "comment cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockCommentRepository)
			uc := usecase.NewCommentUseCase(repo)

			tt.mockSetup(repo)

			err := uc.CreateComment(context.Background(), tt.comment)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestCommentUseCase_GetCommentsByPostID(t *testing.T) {
	tests := []struct {
		name        string
		postID      int
		mockSetup   func(*MockCommentRepository)
		expectedLen int
		expectedErr string
		expectError bool
	}{
		{
			name:   "Success",
			postID: 1,
			mockSetup: func(m *MockCommentRepository) {
				m.On("GetCommentsByPostID", mock.Anything, 1).
					Return([]entity.Comment{
						{ID: 1, Content: "Comment 1"},
						{ID: 2, Content: "Comment 2"},
					}, nil)
			},
			expectedLen: 2,
		},
		{
			name:   "SingleComment",
			postID: 1,
			mockSetup: func(m *MockCommentRepository) {
				m.On("GetCommentsByPostID", mock.Anything, 1).
					Return([]entity.Comment{
						{ID: 1, Content: "Single Comment"},
					}, nil)
			},
			expectedLen: 1,
		},
		{
			name:   "EmptyResult",
			postID: 2,
			mockSetup: func(m *MockCommentRepository) {
				m.On("GetCommentsByPostID", mock.Anything, 2).
					Return([]entity.Comment{}, nil)
			},
			expectedLen: 0,
		},
		{
			name:        "InvalidPostID",
			postID:      0,
			mockSetup:   func(m *MockCommentRepository) {},
			expectedErr: "invalid post ID",
			expectError: true,
		},
		{
			name:        "NegativePostID",
			postID:      -1,
			mockSetup:   func(m *MockCommentRepository) {},
			expectedErr: "invalid post ID",
			expectError: true,
		},
		{
			name:   "RepositoryError",
			postID: 3,
			mockSetup: func(m *MockCommentRepository) {
				m.On("GetCommentsByPostID", mock.Anything, 3).
					Return(nil, errors.New("database error"))
			},
			expectedErr: "database error",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockCommentRepository)
			uc := usecase.NewCommentUseCase(repo)

			tt.mockSetup(repo)

			comments, err := uc.GetCommentsByPostID(context.Background(), tt.postID)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, comments)
			} else {
				require.NoError(t, err)
				assert.Len(t, comments, tt.expectedLen)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestCommentUseCase_DeleteComment(t *testing.T) {
	tests := []struct {
		name        string
		commentID   int
		userID      int
		mockSetup   func(*MockCommentRepository)
		expectedErr string
	}{
		{
			name:      "Success",
			commentID: 1,
			userID:    1,
			mockSetup: func(m *MockCommentRepository) {
				m.On("DeleteComment", mock.Anything, 1, 1).Return(nil)
			},
		},
		{
			name:        "InvalidCommentID",
			commentID:   0,
			userID:      1,
			mockSetup:   func(m *MockCommentRepository) {},
			expectedErr: "invalid comment ID",
		},
		{
			name:        "NegativeCommentID",
			commentID:   -1,
			userID:      1,
			mockSetup:   func(m *MockCommentRepository) {},
			expectedErr: "invalid comment ID",
		},
		{
			name:        "InvalidUserID",
			commentID:   1,
			userID:      0,
			mockSetup:   func(m *MockCommentRepository) {},
			expectedErr: "invalid user ID",
		},
		{
			name:      "RepositoryError",
			commentID: 2,
			userID:    1,
			mockSetup: func(m *MockCommentRepository) {
				m.On("DeleteComment", mock.Anything, 2, 1).
					Return(errors.New("database error"))
			},
			expectedErr: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockCommentRepository)
			uc := usecase.NewCommentUseCase(repo)

			tt.mockSetup(repo)

			err := uc.DeleteComment(context.Background(), tt.commentID, tt.userID)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestNewCommentUseCase(t *testing.T) {
	repo := new(MockCommentRepository)
	uc := usecase.NewCommentUseCase(repo)

	assert.NotNil(t, uc)
	// We can't test the repo field directly since it's unexported
	// Instead we can test behavior by verifying mock calls
	repo.On("CreateComment", mock.Anything, mock.Anything).Return(nil)
	err := uc.CreateComment(context.Background(), &entity.Comment{
		Content: "test",
		PostID:  1,
		UserID:  1,
	})
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}
