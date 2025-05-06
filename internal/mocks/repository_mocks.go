package mocks

import (
	"context"
	"time"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/stretchr/testify/mock"
)

// MockPostgres полностью имитирует repository.Postgres
type MockPostgres struct {
	mock.Mock
	chatRepo *MockChatRepository
}

func NewMockPostgres(chatRepo *MockChatRepository) *MockPostgres {
	return &MockPostgres{
		chatRepo: chatRepo,
	}
}

// Методы для работы с чатом
func (m *MockPostgres) SaveChatMessage(ctx context.Context, msg *entity.ChatMessage) error {
	return m.chatRepo.SaveChatMessage(ctx, msg)
}

func (m *MockPostgres) GetChatMessages(ctx context.Context, limit int) ([]entity.ChatMessage, error) {
	return m.chatRepo.GetChatMessages(ctx, limit)
}

func (m *MockPostgres) DeleteOldChatMessages(ctx context.Context, olderThan time.Duration) error {
	return m.chatRepo.DeleteOldChatMessages(ctx, olderThan)
}

// Остальные методы Postgres (заглушки)
func (m *MockPostgres) CreatePost(ctx context.Context, post *entity.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostgres) GetPostByID(ctx context.Context, id int) (*entity.Post, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Post), args.Error(1)
}

func (m *MockPostgres) GetAllPosts(ctx context.Context) ([]*entity.Post, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Post), args.Error(1)
}

func (m *MockPostgres) DeletePost(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockChatRepository реализует методы для работы с чатом
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

func (m *MockChatRepository) CreateChatMessage(ctx context.Context, message *entity.ChatMessage) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

// MockAuthUseCase для тестирования аутентификации
type MockAuthUseCase struct {
	mock.Mock
}

func (m *MockAuthUseCase) ParseToken(token string) (int64, string, error) {
	args := m.Called(token)
	return args.Get(0).(int64), args.String(1), args.Error(2)
}

// MockPostRepository для тестирования постов
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) CreatePost(ctx context.Context, post *entity.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) GetAllPosts(ctx context.Context) ([]*entity.Post, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Post), args.Error(1)
}

func (m *MockPostRepository) GetPostByID(ctx context.Context, id int) (*entity.Post, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Post), args.Error(1)
}

func (m *MockPostRepository) DeletePost(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockCommentRepository для тестирования комментариев
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

func (m *MockCommentRepository) DeleteComment(ctx context.Context, commentID int, userID int) error {
	args := m.Called(ctx, commentID, userID)
	return args.Error(0)
}

// MockUserRepository для тестирования пользователей
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
