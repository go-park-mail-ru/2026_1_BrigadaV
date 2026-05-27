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

// MeResponse — ответ на GET /api/user/me.
//
//easyjson:json
type MeResponse struct {
	ID        uint64 `json:"id"`
	Login     string `json:"login"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// CSRFTokenResponse — ответ с CSRF токеном.
//
//easyjson:json
type CSRFTokenResponse struct {
	CSRFToken string `json:"csrf_token"`
}

// YandexAuthURLResponse — ответ с URL для авторизации через Яндекс.
//
//easyjson:json
type YandexAuthURLResponse struct {
	URL string `json:"url"`
}
