package models

import "time"

type User struct {
	ID           uint64
	Login        string
	Nickname     string
	AvatarURL    string
	PasswordHash string
	Country      *string
	City         *string
	About        *string
	HasReviews   bool
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
