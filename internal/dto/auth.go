//go:generate easyjson -all auth.go

package dto

//easyjson:json
type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

//easyjson:json
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

//easyjson:json
type LoginResponse struct {
	UserID    uint64 `json:"user_id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}
