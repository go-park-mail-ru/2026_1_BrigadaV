package dto

import "time"

type ProfileResponse struct {
	ID         uint64    `json:"id"`
	Nickname   string    `json:"nickname"`
	Login      string    `json:"login"`
	AvatarURL  string    `json:"avatar_url"`
	Country    *string   `json:"country,omitempty"`
	City       *string   `json:"city,omitempty"`
	About      *string   `json:"about,omitempty"`
	HasReviews bool      `json:"hasReviews"`
	CreatedAt  time.Time `json:"createdAt"`
}

type UpdateProfileRequest struct {
	Nickname  *string `json:"nickname,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Login     *string `json:"login,omitempty"`
	Country   *string `json:"country,omitempty"`
	City      *string `json:"city,omitempty"`
	About     *string `json:"about,omitempty"`
}
