package dto

import "time"

type UserResponse struct {
	ID        uint64    `json:"id"`
	Email     string    `json:"login"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
