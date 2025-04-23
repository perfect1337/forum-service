package entity

import "time"

type ChatMessage struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
