// internal/usecase/post_test.go
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

type MockPostUseCase struct {
	mock.Mock
}
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) CreatePost(ctx context.Context, post *entity.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) GetPostByID(ctx context.Context, id int) (*entity.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Post), args.Error(1)
}

func (m *MockPostRepository) GetAllPosts(ctx context.Context) ([]*entity.Post, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return []*entity.Post{}, args.Error(1) // Возвращаем пустой срез в случае ошибки
	}
	return args.Get(0).([]*entity.Post), args.Error(1)
}
func (m *MockPostRepository) DeletePost(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func TestPostUseCase_CreatePost(t *testing.T) {
	tests := []struct {
		name        string
		post        *entity.Post
		mockSetup   func(*MockPostRepository, *MockUserRepository)
		expectedErr string
	}{
		{
			name: "Success",
			post: &entity.Post{
				Title:   "Test Post",
				Content: "Test Content",
				UserID:  1,
			},
			mockSetup: func(pr *MockPostRepository, ur *MockUserRepository) {
				pr.On("CreatePost", mock.Anything, mock.AnythingOfType("*entity.Post")).Return(nil)
			},
		},
		{
			name: "RepositoryError",
			post: &entity.Post{
				Title:   "Test Post",
				Content: "Test Content",
				UserID:  1,
			},
			mockSetup: func(pr *MockPostRepository, ur *MockUserRepository) {
				pr.On("CreatePost", mock.Anything, mock.AnythingOfType("*entity.Post")).Return(errors.New("database error"))
			},
			expectedErr: "database error",
		},
		{
			name:        "ValidationError_NilPost",
			post:        nil,
			mockSetup:   func(pr *MockPostRepository, ur *MockUserRepository) {},
			expectedErr: "post cannot be nil",
		},
		{
			name: "ValidationError_EmptyTitle",
			post: &entity.Post{
				Title:   "",
				Content: "Content",
				UserID:  1,
			},
			mockSetup:   func(pr *MockPostRepository, ur *MockUserRepository) {},
			expectedErr: "post title cannot be empty",
		},
		{
			name: "ValidationError_EmptyContent",
			post: &entity.Post{
				Title:   "Title",
				Content: "",
				UserID:  1,
			},
			mockSetup:   func(pr *MockPostRepository, ur *MockUserRepository) {},
			expectedErr: "post content cannot be empty",
		},
		{
			name: "ValidationError_EmptyUserID",
			post: &entity.Post{
				Title:   "Title",
				Content: "Content",
				UserID:  0,
			},
			mockSetup:   func(pr *MockPostRepository, ur *MockUserRepository) {},
			expectedErr: "user ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := new(MockPostRepository)
			mockUserRepo := new(MockUserRepository)
			uc := usecase.NewPostUseCase(mockPostRepo, mockUserRepo)

			// Setup mocks
			tt.mockSetup(mockPostRepo, mockUserRepo)

			err := uc.CreatePost(context.Background(), tt.post)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
			}

			mockPostRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestPostUseCase_DeletePost(t *testing.T) {
	tests := []struct {
		name        string
		postID      int
		userID      int
		mockSetup   func(*MockPostRepository, *MockUserRepository)
		expectedErr string
	}{
		{
			name:   "Unauthorized",
			postID: 1,
			userID: 1,
			mockSetup: func(pr *MockPostRepository, ur *MockUserRepository) {
				pr.On("GetPostByID", mock.Anything, 1).Return(&entity.Post{ID: 1, UserID: 2}, nil)
				ur.On("GetUserByID", mock.Anything, 1).Return(&entity.User{ID: 1, Role: "user"}, nil)
			},
			expectedErr: "unauthorized: you can only delete your own posts",
		},
		{
			name:   "SuccessAdmin",
			postID: 1,
			userID: 1,
			mockSetup: func(pr *MockPostRepository, ur *MockUserRepository) {
				pr.On("GetPostByID", mock.Anything, 1).Return(&entity.Post{ID: 1, UserID: 2}, nil)
				ur.On("GetUserByID", mock.Anything, 1).Return(&entity.User{ID: 1, Role: "admin"}, nil)
				pr.On("DeletePost", mock.Anything, 1).Return(nil)
			},
		},
		{
			name:   "SuccessOwner",
			postID: 1,
			userID: 1,
			mockSetup: func(pr *MockPostRepository, ur *MockUserRepository) {
				pr.On("GetPostByID", mock.Anything, 1).Return(&entity.Post{ID: 1, UserID: 1}, nil)
				ur.On("GetUserByID", mock.Anything, 1).Return(&entity.User{ID: 1, Role: "user"}, nil)
				pr.On("DeletePost", mock.Anything, 1).Return(nil)
			},
		},
		{
			name:   "PostNotFound",
			postID: 1,
			userID: 1,
			mockSetup: func(pr *MockPostRepository, ur *MockUserRepository) {
				pr.On("GetPostByID", mock.Anything, 1).Return(nil, errors.New("post not found"))
			},
			expectedErr: "post not found",
		},
		{
			name:   "UserNotFound",
			postID: 1,
			userID: 1,
			mockSetup: func(pr *MockPostRepository, ur *MockUserRepository) {
				pr.On("GetPostByID", mock.Anything, 1).Return(&entity.Post{ID: 1, UserID: 1}, nil)
				ur.On("GetUserByID", mock.Anything, 1).Return(nil, errors.New("user not found"))
			},
			expectedErr: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := new(MockPostRepository)
			mockUserRepo := new(MockUserRepository)
			uc := usecase.NewPostUseCase(mockPostRepo, mockUserRepo)

			// Setup mocks
			tt.mockSetup(mockPostRepo, mockUserRepo)

			err := uc.DeletePost(context.Background(), tt.postID, tt.userID)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
			}

			mockPostRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}
func TestPostUseCase_GetPostByID(t *testing.T) {
	tests := []struct {
		name         string
		postID       int
		mockSetup    func(*MockPostRepository)
		expectedErr  string
		expectedPost *entity.Post
	}{
		{
			name:   "Success",
			postID: 1,
			mockSetup: func(pr *MockPostRepository) {
				pr.On("GetPostByID", mock.Anything, 1).Return(&entity.Post{ID: 1, Title: "Test Post", Content: "Test Content", UserID: 1}, nil)
			},
			expectedPost: &entity.Post{ID: 1, Title: "Test Post", Content: "Test Content", UserID: 1},
		},
		{
			name:   "PostNotFound",
			postID: 1,
			mockSetup: func(pr *MockPostRepository) {
				pr.On("GetPostByID", mock.Anything, 1).Return(nil, errors.New("post not found"))
			},
			expectedErr: "post not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := new(MockPostRepository)
			mockUserRepo := new(MockUserRepository)
			uc := usecase.NewPostUseCase(mockPostRepo, mockUserRepo)

			// Setup mocks
			tt.mockSetup(mockPostRepo)

			post, err := uc.GetPostByID(context.Background(), tt.postID)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPost, post)
			}

			mockPostRepo.AssertExpectations(t)
		})
	}
}
func TestPostUseCase_GetAllPosts(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockPostRepository)
		expectedErr   string
		expectedPosts []*entity.Post
	}{
		{
			name: "Success",
			mockSetup: func(pr *MockPostRepository) {
				pr.On("GetAllPosts", mock.Anything).Return([]*entity.Post{
					{ID: 1, Title: "Test Post 1", Content: "Test Content 1", UserID: 1},
					{ID: 2, Title: "Test Post 2", Content: "Test Content 2", UserID: 2},
				}, nil)
			},
			expectedPosts: []*entity.Post{
				{ID: 1, Title: "Test Post 1", Content: "Test Content 1", UserID: 1},
				{ID: 2, Title: "Test Post 2", Content: "Test Content 2", UserID: 2},
			},
		},
		{
			name: "RepositoryError",
			mockSetup: func(pr *MockPostRepository) {
				pr.On("GetAllPosts", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedErr: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := new(MockPostRepository)
			mockUserRepo := new(MockUserRepository)
			uc := usecase.NewPostUseCase(mockPostRepo, mockUserRepo)

			// Setup mocks
			tt.mockSetup(mockPostRepo)

			posts, err := uc.GetAllPosts(context.Background())

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPosts, posts)
			}

			mockPostRepo.AssertExpectations(t)
		})
	}
}
