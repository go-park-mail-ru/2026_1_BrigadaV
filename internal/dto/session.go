package dto

type SessionResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}
