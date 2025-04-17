package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/perfect1337/forum-service/internal/config"
	"github.com/perfect1337/forum-service/internal/entity"
	"golang.org/x/crypto/bcrypt"
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

func (p *Postgres) CreateUser(ctx context.Context, user *entity.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = string(hashedPassword)

	query := `INSERT INTO users (username, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id`
	return p.db.QueryRowContext(ctx, query, user.Username, user.Email, user.PasswordHash, user.Role).Scan(&user.ID)
}

func (p *Postgres) GetUserByCredentials(ctx context.Context, email, password string) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, username, email, password_hash, role FROM users WHERE email = $1`
	err := p.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return &user, nil
}

func (p *Postgres) CreateRefreshToken(ctx context.Context, token *entity.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`
	_, err := p.db.ExecContext(ctx, query, token.UserID, token.Token, token.ExpiresAt)
	return err
}

func (p *Postgres) GetRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	var rt entity.RefreshToken
	query := `SELECT id, user_id, token, expires_at FROM refresh_tokens WHERE token = $1`
	err := p.db.QueryRowContext(ctx, query, token).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("refresh token not found")
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	return &rt, nil
}

func (p *Postgres) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := p.db.ExecContext(ctx, query, token)
	return err
}
