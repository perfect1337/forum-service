package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq" // Драйвер для PostgreSQL
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/stretchr/testify/assert"
)

func setupTestDB() (*Postgres, error) {
	// Замените на строку подключения к вашей тестовой базе данных
	connStr := "user=postgres dbname=PG sslmode=disable password=postgres"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть базу данных: %v", err)
	}

	// Создание тестовой таблицы
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS chat_messages (
			id SERIAL PRIMARY KEY,
			user_id INTEGER,
			author VARCHAR(255),
			text TEXT,
			created_at TIMESTAMP
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать тестовую таблицу: %v", err)
	}

	return &Postgres{db: db}, nil
}

func TestPostgresCreateChatMessage(t *testing.T) {
	repo, err := setupTestDB()
	if err != nil {
		t.Fatalf("не удалось настроить тестовую базу данных: %v", err)
	}

	ctx := context.Background()
	message := &entity.ChatMessage{
		UserID: 1,
		Author: "testuser",
		Text:   "Hello, World!",
	}

	err = repo.CreateChatMessage(ctx, message)
	assert.NoError(t, err)
	assert.NotZero(t, message.ID)
}

func TestPostgresGetChatMessages(t *testing.T) {
	repo, err := setupTestDB()
	if err != nil {
		t.Fatalf("не удалось настроить тестовую базу данных: %v", err)
	}

	ctx := context.Background()
	limit := 10

	// Вставка тестовых данных
	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO chat_messages (user_id, author, text, created_at)
		VALUES (1, 'testuser', 'Hello, World!', NOW())
	`)
	if err != nil {
		t.Fatalf("не удалось вставить тестовые данные: %v", err)
	}

	messages, err := repo.GetChatMessages(ctx, limit)
	assert.NoError(t, err)
	assert.NotEmpty(t, messages)
}

func TestPostgresSaveChatMessage(t *testing.T) {
	repo, err := setupTestDB()
	if err != nil {
		t.Fatalf("не удалось настроить тестовую базу данных: %v", err)
	}

	ctx := context.Background()
	message := &entity.ChatMessage{
		UserID: 1,
		Author: "testuser",
		Text:   "Hello, World!",
	}

	err = repo.SaveChatMessage(ctx, message)
	assert.NoError(t, err)
	assert.NotZero(t, message.ID)
	assert.NotZero(t, message.CreatedAt)
}

func TestPostgresDeleteOldChatMessages(t *testing.T) {
	repo, err := setupTestDB()
	if err != nil {
		t.Fatalf("не удалось настроить тестовую базу данных: %v", err)
	}

	ctx := context.Background()
	olderThan := time.Hour * 24

	// Вставка тестовых данных
	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO chat_messages (user_id, author, text, created_at)
		VALUES (1, 'testuser', 'Hello, World!', NOW() - INTERVAL '25 hours')
	`)
	if err != nil {
		t.Fatalf("не удалось вставить тестовые данные: %v", err)
	}

	err = repo.DeleteOldChatMessages(ctx, olderThan)
	assert.NoError(t, err)
}
