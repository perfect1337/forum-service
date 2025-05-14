package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresGetUserByID(t *testing.T) {
	repo, err := setupTestDB()
	require.NoError(t, err, "Failed to setup test database")

	ctx := context.Background()

	// Генерируем уникальные тестовые данные
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("user_%d", timestamp)
	email := fmt.Sprintf("user_%d@example.com", timestamp)
	role := "user"

	// Вставляем тестового пользователя
	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO users (username, email, password_hash, role)
		VALUES ($1, $2, 'hashed_password', $3)
	`, username, email, role)
	require.NoError(t, err, "Failed to insert test user")

	// Получаем ID созданного пользователя
	var userID int
	err = repo.db.QueryRowContext(ctx, "SELECT id FROM users WHERE username = $1", username).Scan(&userID)
	require.NoError(t, err, "Failed to get user ID")

	// Тестируем получение пользователя
	t.Run("Success", func(t *testing.T) {
		user, err := repo.GetUserByID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, username, user.Username)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, role, user.Role)
	})

	t.Run("Not Found", func(t *testing.T) {
		_, err := repo.GetUserByID(ctx, 99999)
		assert.Error(t, err)
		assert.Equal(t, "sql: no rows in result set", err.Error())
	})
}

func TestPostgresGetUsersByIDs(t *testing.T) {
	repo, err := setupTestDB()
	require.NoError(t, err, "Failed to setup test database")

	ctx := context.Background()

	// Генерируем тестовых пользователей
	timestamp := time.Now().UnixNano()
	usersData := []struct {
		username string
		email    string
		role     string
	}{
		{fmt.Sprintf("user1_%d", timestamp), fmt.Sprintf("user1_%d@example.com", timestamp), "user"},
		{fmt.Sprintf("user2_%d", timestamp), fmt.Sprintf("user2_%d@example.com", timestamp), "admin"},
	}

	var userIDs []int
	// Вставляем тестовых пользователей
	for _, data := range usersData {
		_, err = repo.db.ExecContext(ctx, `
            INSERT INTO users (username, email, password_hash, role)
            VALUES ($1, $2, 'hashed_password', $3)
        `, data.username, data.email, data.role)
		require.NoError(t, err, "Failed to insert test user")

		var userID int
		err = repo.db.QueryRowContext(ctx, "SELECT id FROM users WHERE username = $1", data.username).Scan(&userID)
		require.NoError(t, err, "Failed to get user ID")
		userIDs = append(userIDs, userID)
	}

	t.Run("Success with multiple users", func(t *testing.T) {
		users, err := repo.GetUsersByIDs(ctx, userIDs)
		require.NoError(t, err)
		assert.Len(t, users, 2)

		for id, user := range users {
			assert.Contains(t, userIDs, id)
			assert.NotEmpty(t, user.Username)
			assert.NotEmpty(t, user.Email)
		}
	})

	t.Run("Success with single user", func(t *testing.T) {
		users, err := repo.GetUsersByIDs(ctx, []int{userIDs[0]})
		require.NoError(t, err)
		assert.Len(t, users, 1)
	})

	t.Run("Empty result for non-existent IDs", func(t *testing.T) {
		users, err := repo.GetUsersByIDs(ctx, []int{99999})
		require.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("Mixed existing and non-existent IDs", func(t *testing.T) {
		ids := append(userIDs, 99999)
		users, err := repo.GetUsersByIDs(ctx, ids)
		require.NoError(t, err)
		assert.Len(t, users, 2) // Только существующие пользователи
	})

	t.Run("Empty IDs slice", func(t *testing.T) {
		users, err := repo.GetUsersByIDs(ctx, []int{})
		require.NoError(t, err)
		assert.Empty(t, users)
	})
}

func TestPostgresUserRepositoryEdgeCases(t *testing.T) {
	repo, err := setupTestDB()
	require.NoError(t, err, "Failed to setup test database")

	ctx := context.Background()

	t.Run("Empty IDs slice", func(t *testing.T) {
		users, err := repo.GetUsersByIDs(ctx, []int{})
		require.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Немедленно отменяем контекст

		_, err := repo.GetUserByID(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
