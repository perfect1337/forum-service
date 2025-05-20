package repository

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/perfect1337/forum-service/internal/config"
	"github.com/perfect1337/forum-service/internal/entity"
)

type PostRepository interface {
	CreatePost(ctx context.Context, post *entity.Post) error
	GetAllPosts(ctx context.Context) ([]*entity.Post, error)
	GetPostByID(ctx context.Context, id int) (*entity.Post, error)
	DeletePost(ctx context.Context, id int) error
	UpdatePost(ctx context.Context, postID int, title, content string) error
}

type Postgres struct {
	db  *sql.DB
	cfg *config.Config
}

func NewPostgres(cfg *config.Config) (*Postgres, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User,
		cfg.Postgres.Password, cfg.Postgres.DBName, cfg.Postgres.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return &Postgres{db: db, cfg: cfg}, nil
}

func (p *Postgres) CreatePost(ctx context.Context, post *entity.Post) error {
	query := `INSERT INTO posts (title, content, user_id) VALUES ($1, $2, $3) 
              RETURNING id, created_at`
	return p.db.QueryRowContext(ctx, query, post.Title, post.Content, post.UserID).
		Scan(&post.ID, &post.CreatedAt)
}

func (p *Postgres) GetAllPosts(ctx context.Context) ([]*entity.Post, error) {
	query := `
            SELECT 
            p.id, 
            p.title, 
            p.content, 
            p.user_id, 
            u.username AS author,  -- Получаем имя автора из users
            p.created_at
        FROM posts p
        JOIN users u ON p.user_id = u.id  -- Важно: соединяем с таблицей users
        ORDER BY p.created_at DESC
    `
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()

	var posts []*entity.Post
	for rows.Next() {
		var post entity.Post
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserID,
			&post.Author, // Получаем username из таблицы users
			&post.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, &post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return posts, nil
}
func (p *Postgres) GetPostByID(ctx context.Context, id int) (*entity.Post, error) {
	query := `
        SELECT p.id, p.title, p.content, p.user_id, u.username, p.created_at
        FROM posts p
        JOIN users u ON p.user_id = u.id
        WHERE p.id = $1
    `
	var post entity.Post
	err := p.db.QueryRowContext(ctx, query, id).
		Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserID,
			&post.Author,
			&post.CreatedAt,
		)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by ID: %w", err)
	}
	return &post, nil
}

func (p *Postgres) DeletePost(ctx context.Context, id int) error {
	query := `DELETE FROM posts WHERE id = $1`
	_, err := p.db.ExecContext(ctx, query, id)
	return err
}

func (p *Postgres) UpdatePost(ctx context.Context, postID int, title, content string) error {
	query := `UPDATE posts SET title = $1, content = $2 WHERE id = $3`
	_, err := p.db.ExecContext(ctx, query, title, content, postID)
	return err
}
