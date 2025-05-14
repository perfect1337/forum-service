package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/perfect1337/forum-service/internal/entity"
)

type ChatRepository interface {
	CreateChatMessage(ctx context.Context, message *entity.ChatMessage) error
	GetChatMessages(ctx context.Context, limit int) ([]entity.ChatMessage, error)
	SaveChatMessage(ctx context.Context, message *entity.ChatMessage) error
	DeleteOldChatMessages(ctx context.Context, olderThan time.Duration) error
}

func (p *Postgres) DeleteOldChatMessages(ctx context.Context, olderThan time.Duration) error {
	query := `DELETE FROM chat_messages WHERE created_at < NOW() - $1::interval`
	_, err := p.db.ExecContext(ctx, query, olderThan.String())
	return err
}

func (p *Postgres) CreateChatMessage(ctx context.Context, message *entity.ChatMessage) error {
	query := `INSERT INTO chat_messages (user_id, author, text, created_at) 
              VALUES ($1, $2, $3, $4) RETURNING id`
	return p.db.QueryRowContext(ctx, query,
		message.UserID,
		message.Author,
		message.Text,
		time.Now(),
	).Scan(&message.ID)
}

func (p *Postgres) GetChatMessages(ctx context.Context, limit int) ([]entity.ChatMessage, error) {
	query := `
        SELECT id, user_id, author, text, created_at 
        FROM chat_messages 
        ORDER BY created_at DESC 
        LIMIT $1
    `

	rows, err := p.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat messages: %w", err)
	}
	defer rows.Close()

	var messages []entity.ChatMessage
	for rows.Next() {
		var msg entity.ChatMessage
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.Author, &msg.Text, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan chat message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
func (p *Postgres) SaveChatMessage(ctx context.Context, message *entity.ChatMessage) error {
	query := `
        INSERT INTO chat_messages (user_id, author, text, created_at)
        VALUES ($1, $2, $3, NOW())
        RETURNING id, created_at
    `

	err := p.db.QueryRowContext(
		ctx,
		query,
		message.UserID,
		message.Author,
		message.Text,
	).Scan(&message.ID, &message.CreatedAt)

	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}
