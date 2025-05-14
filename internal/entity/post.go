package entity

import "time"

// internal/entity/post.go
type Post struct {
	ID        int       `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	UserID    int       `json:"user_id" db:"user_id"`
	Author    string    `json:"author" db:"-"` // db:"-" означает, что это поле не маппится напрямую
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Comments  []Comment `json:"comments,omitempty" db:"-"`
}
type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type RefreshToken struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
