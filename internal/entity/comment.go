package entity

import "time"

type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	AuthorID  string    `json:"author_id"`
	Author    string    `json:"author"` // Может заполняться при получении
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
