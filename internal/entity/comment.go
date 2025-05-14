package entity

import "time"

type Comment struct {
	ID        int       `json:"id" db:"id"`
	Content   string    `json:"content" db:"content"`
	PostID    int       `json:"post_id" db:"post_id"` // Должно быть int
	UserID    int       `json:"user_id" db:"user_id"`
	Author    string    `json:"author" db:"-"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
