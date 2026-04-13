package models

import "time"

type User struct {
	ID           uint64
	Nickname     string
	AvatarURL    string
	PasswordHash string
	Country      *string
	City         *string
	About        *string
	HasReviews   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
