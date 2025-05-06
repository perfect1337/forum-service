// internal/mocks/mocks.go
package mocks

import (
	"context"
	"time"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/proto/user"
	"github.com/perfect1337/forum-service/internal/usecase"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// MockPostgres имитирует repository.Postgres
type MockPostgres struct {
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
	return args.Get(0).(*entity.Post), args.Error(1)
}

func (m *MockPostRepository) GetAllPosts(ctx context.Context) ([]*entity.Post, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Post), args.Error(1)
}

func (m *MockPostRepository) DeletePost(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockPostUseCase реализует usecase.PostUseCase
var _ usecase.PostUseCase = (*MockPostUseCase)(nil)

type MockPostUseCase struct {
	mock.Mock
}

func (m *MockPostUseCase) CreatePost(ctx context.Context, post *entity.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostUseCase) GetPostByID(ctx context.Context, id int) (*entity.Post, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Post), args.Error(1)
}

func (m *MockPostUseCase) GetAllPosts(ctx context.Context) ([]*entity.Post, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Post), args.Error(1)
}

func (m *MockPostUseCase) DeletePost(ctx context.Context, postID, userID int) error {
	args := m.Called(ctx, postID, userID)
	return args.Error(0)
}

// MockUserServiceClient имитирует UserServiceClient
type MockUserServiceClient struct {
	mock.Mock
}

func (m *MockUserServiceClient) GetUsername(ctx context.Context, in *user.UserRequest, opts ...grpc.CallOption) (*user.UserResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*user.UserResponse), args.Error(1)
}

// MockAuthUseCase имитирует usecase.AuthUseCase
type MockAuthUseCase struct {
	mock.Mock
}

func (m *MockAuthUseCase) ParseToken(token string) (int64, string, error) {
	args := m.Called(token)
	return args.Get(0).(int64), args.String(1), args.Error(2)
}

// MockChatRepository имитирует repository.ChatRepository
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

// MockCommentRepository имитирует repository.CommentRepository
type MockCommentRepository struct {
	mock.Mock
}

func (m *MockCommentRepository) CreateComment(ctx context.Context, comment *entity.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentRepository) GetCommentsByPostID(ctx context.Context, postID int) ([]entity.Comment, error) {
	args := m.Called(ctx, postID)
	return args.Get(0).([]entity.Comment), args.Error(1)
}

func (m *MockCommentRepository) DeleteComment(ctx context.Context, commentID, userID int) error {
	args := m.Called(ctx, commentID, userID)
	return args.Error(0)
}

// MockUserRepository имитирует repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetUsersByIDs(ctx context.Context, ids []int) (map[int]*entity.User, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(map[int]*entity.User), args.Error(1)
}
