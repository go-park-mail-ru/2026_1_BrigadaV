package middleware

import (
	"context"
	"encoding/json"
	"guidely-app/internal/repository"
	"net/http"
	"time"
)

type AuthMiddleware struct {
	sessionRepo *repository.SessionRepo
}

func NewAuthMiddleware(sessionRepo *repository.SessionRepo) *AuthMiddleware {
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
