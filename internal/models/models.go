package models

import "time"

type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
}

type Session struct {
	HashedToken string    `json:"-"`
	UserID      int       `json:"-"`
	ExpiresAt   time.Time `json:"-"`
}

type Chat struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	OwnerID int    `json:"owner_id"`
}

type ChatMember struct {
	UserID int `json:"user_id"`
	ChatID int `json:"chat_id"`
}

type Message struct {
	ID        int        `json:"id"`
	ChatID    int        `json:"chat_id"`
	UserID    int        `json:"user_id"`
	Text      string     `json:"text"`
	CreatedAt time.Time  `json:"created_at"`
	EditedAt  *time.Time `json:"edited_at"`
}
