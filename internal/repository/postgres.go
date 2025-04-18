package repository

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/perfect1337/forum-service/internal/config"
	"github.com/perfect1337/forum-service/internal/entity"
)

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
	query := `INSERT INTO posts (title, content, author) VALUES ($1, $2, $3) RETURNING id, created_at`
	return p.db.QueryRowContext(ctx, query, post.Title, post.Content, post.Author).Scan(&post.ID, &post.CreatedAt)
}

func (p *Postgres) GetAllPosts(ctx context.Context) ([]*entity.Post, error) {
	query := `SELECT id, title, content, author, created_at FROM posts ORDER BY created_at DESC`
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*entity.Post
	for rows.Next() {
		var post entity.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (p *Postgres) GetPostByID(ctx context.Context, id int) (*entity.Post, error) {
	query := `SELECT id, title, content, author, created_at FROM posts WHERE id = $1`
	var post entity.Post
	err := p.db.QueryRowContext(ctx, query, id).Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &post, nil
}
