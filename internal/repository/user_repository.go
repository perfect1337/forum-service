package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/stretchr/testify/mock"
)

var (
	ErrNotFound = errors.New("not found")
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id int) (*entity.User, error)
	GetUsersByIDs(ctx context.Context, ids []int) (map[int]*entity.User, error)
	CreateUser(ctx context.Context, user *entity.User) error
}
type MockUserRepository struct {
	mock.Mock
}

func (p *Postgres) CreateUser(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, username, email, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := p.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.Role,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}
func (p *Postgres) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	query := `SELECT id, username, email, role FROM users WHERE id = $1`
	var user entity.User
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (p *Postgres) GetUsersByIDs(ctx context.Context, ids []int) (map[int]*entity.User, error) {
	if len(ids) == 0 {
		return make(map[int]*entity.User), nil
	}

	query := `SELECT id, username, email, role FROM users WHERE id = ANY($1)`
	rows, err := p.db.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make(map[int]*entity.User)
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
		); err != nil {
			return nil, err
		}
		users[user.ID] = &user
	}
	return users, nil
}
