package dto

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UserID    uint64 `json:"user_id"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}
