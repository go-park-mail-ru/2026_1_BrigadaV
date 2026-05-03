package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	authrepo "guidely-app/internal/auth/repository"
)

type AuthMiddleware struct {
	sessionRepo authrepo.SessionRepository
}

func NewAuthMiddleware(sessionRepo authrepo.SessionRepository) *AuthMiddleware {
	return &AuthMiddleware{sessionRepo: sessionRepo}
}

func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized", "message": "missing session cookie"})
			return
		}
		session, err := m.sessionRepo.GetByToken(r.Context(), cookie.Value)
		if err != nil || session == nil || session.ExpiresAt.Before(time.Now()) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized", "message": "invalid or expired session"})
			return
		}
		ctx := context.WithValue(r.Context(), "user_id", session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GetUserIDFromContext(r *http.Request) uint64 {
	val := r.Context().Value("user_id")
	if id, ok := val.(uint64); ok {
		return id
	}
	return 0
}
