package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/perfect1337/forum-service/internal/entity"
)

func (r *Postgres) CreateComment(ctx context.Context, comment *entity.Comment) error {
	query := `
        INSERT INTO comments (post_id, author_id, author, text, created_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `

	err := r.db.QueryRowContext(ctx, query,
		comment.PostID,
		comment.AuthorID,
		comment.Author, // Добавляем имя автора
		comment.Text,
		comment.CreatedAt,
	).Scan(&comment.ID)

	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}
func (r *Postgres) GetCommentsByPostID(ctx context.Context, postID int) ([]entity.Comment, error) {
	query := `
        SELECT c.id, c.post_id, c.author_id, u.username as author, c.text, c.created_at
        FROM comments c
        JOIN users u ON c.author_id = u.id::text  -- Приводим id к text если необходимо
        WHERE c.post_id = $1
        ORDER BY c.created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	defer rows.Close()

	var comments []entity.Comment
	for rows.Next() {
		var comment entity.Comment
		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.AuthorID,
			&comment.Author,
			&comment.Text,
			&comment.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}
func (p *Postgres) DeleteComment(ctx context.Context, commentID int, userID string) error {
	query := `DELETE FROM comments WHERE id = $1 AND author_id = $2`

	result, err := p.db.ExecContext(ctx, query, commentID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
