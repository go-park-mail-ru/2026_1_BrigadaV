package dto

import "time"

type ProfileResponse struct {
	ID        uint64    `json:"id"`
	Login     string    `json:"login"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateProfileRequest struct {
	Nickname    string `json:"nickname,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	OldPassword string `json:"old_password,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
}
