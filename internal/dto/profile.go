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
    Nickname  string `json:"nickname"`
    AvatarURL string `json:"avatar_url"`
}

type ChangePasswordRequest struct {
    OldPassword string `json:"old_password"`
    NewPassword string `json:"new_password"`
}