package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	authrepo "guidely-app/internal/auth/repository"
	"guidely-app/pkg/utils"
)

type contextKey string

const UserIDKey contextKey = "user_id"

type AuthMiddleware struct {
	sessionRepo authrepo.SessionRepository
}

func NewAuthMiddleware(sessionRepo authrepo.SessionRepository) *AuthMiddleware {
	return &AuthMiddleware{sessionRepo: sessionRepo}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			log.Printf("Auth: missing session_token cookie for %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized", "message": "missing session cookie"})
			return
		}
		hashed := utils.HashToken(cookie.Value)
		session, err := m.sessionRepo.GetByToken(r.Context(), hashed)
		if err != nil {
			log.Printf("Auth: GetByToken error: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized", "message": "database error"})
			return
		}
		if session == nil || session.ExpiresAt.Before(time.Now()) {
			log.Printf("Auth: session not found or expired for token hash %s", hashed[:10])
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized", "message": "invalid or expired session"})
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(r *http.Request) uint64 {
	val := r.Context().Value(UserIDKey)
	if id, ok := val.(uint64); ok {
		return id
	}
	return 0
}
