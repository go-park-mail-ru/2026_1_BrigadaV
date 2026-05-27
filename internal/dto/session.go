//go:generate easyjson -all session.go

package dto

//easyjson:json
type SessionResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}
