package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq" // Драйвер для PostgreSQL
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresCreatePost(t *testing.T) {
	repo, err := setupTestDB()
	if err != nil {
		t.Fatalf("не удалось настроить тестовую базу данных: %v", err)
	}

	ctx := context.Background()
	post := &entity.Post{
		Title:   "Test Post",
		Content: "This is a test post",
		UserID:  1,
	}

	err = repo.CreatePost(ctx, post)
	assert.NoError(t, err)
	assert.NotZero(t, post.ID)
}

func TestPostgresGetAllPosts(t *testing.T) {
	repo, err := setupTestDB()
	require.NoError(t, err, "Failed to setup test database")

	ctx := context.Background()

	// Генерируем уникальные данные для каждого теста
	timestamp := time.Now().Unix()
	username := fmt.Sprintf("testuser_%d", timestamp)
	email := fmt.Sprintf("testuser_%d@example.com", timestamp)

	// Вставляем тестового пользователя
	_, err = repo.db.ExecContext(ctx, `
        INSERT INTO users (username, email, password_hash)
        VALUES ($1, $2, 'hashedpassword')
    `, username, email)
	require.NoError(t, err, "Failed to insert test user")

	// Получаем ID созданного пользователя
	var userID int
	err = repo.db.QueryRowContext(ctx, "SELECT id FROM users WHERE username = $1", username).Scan(&userID)
	require.NoError(t, err, "Failed to get user ID")

	// Вставляем тестовый пост
	_, err = repo.db.ExecContext(ctx, `
        INSERT INTO posts (title, content, user_id, created_at)
        VALUES ('Test Post', 'This is a test post', $1, NOW())
    `, userID)
	require.NoError(t, err, "Failed to insert test post")

	// Тестируем получение постов
	posts, err := repo.GetAllPosts(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, posts)
	assert.Equal(t, "Test Post", posts[0].Title)
	assert.Equal(t, username, posts[0].Author)
}

func TestPostgresGetPostByID(t *testing.T) {
	repo, err := setupTestDB()
	require.NoError(t, err)

	ctx := context.Background()

	// Генерируем уникальные данные для теста
	timestamp := time.Now().UnixNano() // Используем UnixNano для большей уникальности
	username := fmt.Sprintf("testuser_%d", timestamp)
	email := fmt.Sprintf("testuser_%d@example.com", timestamp)

	// Вставляем тестового пользователя
	_, err = repo.db.ExecContext(ctx, `
        INSERT INTO users (username, email, password_hash)
        VALUES ($1, $2, 'hash')
    `, username, email)
	require.NoError(t, err, "Failed to insert test user")

	// Получаем ID созданного пользователя
	var userID int
	err = repo.db.QueryRowContext(ctx, "SELECT id FROM users WHERE username = $1", username).Scan(&userID)
	require.NoError(t, err, "Failed to get user ID")

	// Вставляем тестовый пост
	postTitle := fmt.Sprintf("Test Post %d", timestamp)
	_, err = repo.db.ExecContext(ctx, `
        INSERT INTO posts (title, content, user_id)
        VALUES ($1, 'Content', $2)
    `, postTitle, userID)
	require.NoError(t, err, "Failed to insert test post")

	// Получаем ID созданного поста
	var postID int
	err = repo.db.QueryRowContext(ctx, "SELECT id FROM posts WHERE title = $1", postTitle).Scan(&postID)
	require.NoError(t, err, "Failed to get post ID")

	// Тестируем получение поста по ID
	post, err := repo.GetPostByID(ctx, postID)
	require.NoError(t, err)
	assert.Equal(t, postTitle, post.Title)
	assert.Equal(t, username, post.Author)
}
func TestPostgresDeletePost(t *testing.T) {
	repo, err := setupTestDB()
	require.NoError(t, err, "Failed to setup test database")

	ctx := context.Background()

	// Генерируем уникальные данные для пользователя
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("user_%d", timestamp)
	email := fmt.Sprintf("test_%d@example.com", timestamp) // Уникальный email

	// Создаем тестового пользователя
	_, err = repo.db.ExecContext(ctx, `
        INSERT INTO users (username, email, password_hash)
        VALUES ($1, $2, $3)
    `, username, email, "hash")
	require.NoError(t, err, "Failed to insert test user")

	// Получаем ID созданного пользователя
	var userID int
	err = repo.db.QueryRowContext(ctx, "SELECT id FROM users WHERE username = $1", username).Scan(&userID)
	require.NoError(t, err, "Failed to get user ID")

	// Создаем тестовый пост
	postTitle := fmt.Sprintf("Test Post %d", timestamp)
	_, err = repo.db.ExecContext(ctx, `
        INSERT INTO posts (title, content, user_id)
        VALUES ($1, 'This is a test post', $2)
    `, postTitle, userID)
	require.NoError(t, err, "Failed to insert test post")

	// Получаем ID созданного поста
	var postID int
	err = repo.db.QueryRowContext(ctx, "SELECT id FROM posts WHERE title = $1", postTitle).Scan(&postID)
	require.NoError(t, err, "Failed to get post ID")

	// Тестируем удаление
	err = repo.DeletePost(ctx, postID)
	assert.NoError(t, err)

	// Проверяем, что пост действительно удален
	var count int
	err = repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM posts WHERE id = $1", postID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
