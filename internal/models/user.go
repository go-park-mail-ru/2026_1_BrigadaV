package models

import "time"

type User struct {
	ID           uint64
	Email        string
	Nickname     string
	AvatarURL    string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
