package repository

import (
	"context"
	"testing"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCommentRepository is a mock implementation of CommentRepository
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

func TestPostgresCreateComment(t *testing.T) {
	repo, err := setupTestDB()
	if err != nil {
		t.Fatalf("не удалось настроить тестовую базу данных: %v", err)
	}

	ctx := context.Background()
	comment := &entity.Comment{
		Content: "Test comment",
		PostID:  1,
		UserID:  1,
	}

	err = repo.CreateComment(ctx, comment)
	assert.NoError(t, err)
	assert.NotZero(t, comment.ID)
}

func TestPostgresGetCommentsByPostID(t *testing.T) {
	repo, err := setupTestDB()
	if err != nil {
		t.Fatalf("не удалось настроить тестовую базу данных: %v", err)
	}

	ctx := context.Background()
	postID := 1

	// Удаляем существующие записи перед вставкой
	_, err = repo.db.ExecContext(ctx, `
		DELETE FROM users WHERE id = 1
	`)
	if err != nil {
		t.Fatalf("не удалось удалить тестовые данные: %v", err)
	}

	// Вставка тестовых данных
	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO users (id, username, email, password_hash) VALUES (1, 'testuser', 'testuser@example.com', 'hashedpassword')
	`)
	if err != nil {
		t.Fatalf("не удалось вставить тестовые данные: %v", err)
	}

	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO comments (content, post_id, user_id, created_at)
		VALUES ('Test comment', 1, 1, NOW())
	`)
	if err != nil {
		t.Fatalf("не удалось вставить тестовые данные: %v", err)
	}

	comments, err := repo.GetCommentsByPostID(ctx, postID)
	assert.NoError(t, err)
	assert.NotEmpty(t, comments)
}
func TestPostgresDeleteComment(t *testing.T) {
	repo, err := setupTestDB()
	if err != nil {
		t.Fatalf("не удалось настроить тестовую базу данных: %v", err)
	}

	ctx := context.Background()
	commentID := 1
	userID := 1

	// Удаляем существующие записи перед вставкой
	_, err = repo.db.ExecContext(ctx, `
		DELETE FROM comments WHERE id = 1
	`)
	if err != nil {
		t.Fatalf("не удалось удалить тестовые данные: %v", err)
	}

	// Вставка тестовых данных
	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO comments (id, content, post_id, user_id, created_at)
		VALUES (1, 'Test comment', 1, 1, NOW())
	`)
	if err != nil {
		t.Fatalf("не удалось вставить тестовые данные: %v", err)
	}

	err = repo.DeleteComment(ctx, commentID, userID)
	assert.NoError(t, err)
}
