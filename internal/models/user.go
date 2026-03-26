package models

import "time"

type User struct {
	ID           uint64    `json:"id"`
	Email        string    `json:"email"`
	Nickname     string    `json:"nickname"`
	AvatarURL    string    `json:"avatar_url"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
