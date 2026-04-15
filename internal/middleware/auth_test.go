package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"guidely-app/internal/models"
	"guidely-app/internal/repository/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_Authenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	authMiddleware := NewAuthMiddleware(mockSessionRepo)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("authenticated"))
		_ = userID
	})

	tests := []struct {
		name           string
		cookie         *http.Cookie
		mockBehavior   func()
		expectedStatus int
		checkContext   bool
	}{
		{
			name:   "missing cookie",
			cookie: nil,
			mockBehavior: func() {
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid token",
			cookie: &http.Cookie{
				Name:  "session_token",
				Value: "invalid",
			},
			mockBehavior: func() {
				mockSessionRepo.EXPECT().GetByToken(gomock.Any(), "invalid").Return(nil, nil)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "expired session",
			cookie: &http.Cookie{
				Name:  "session_token",
				Value: "expired_token",
			},
			mockBehavior: func() {
				session := &models.Session{
					UserID:    1,
					ExpiresAt: time.Now().Add(-1 * time.Hour),
				}
				mockSessionRepo.EXPECT().GetByToken(gomock.Any(), "expired_token").Return(session, nil)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid session",
			cookie: &http.Cookie{
				Name:  "session_token",
				Value: "valid_token",
			},
			mockBehavior: func() {
				session := &models.Session{
					UserID:    42,
					ExpiresAt: time.Now().Add(1 * time.Hour),
				}
				mockSessionRepo.EXPECT().GetByToken(gomock.Any(), "valid_token").Return(session, nil)
			},
			expectedStatus: http.StatusOK,
			checkContext:   true,
		},
		{
			name: "repository error",
			cookie: &http.Cookie{
				Name:  "session_token",
				Value: "error_token",
			},
			mockBehavior: func() {
				mockSessionRepo.EXPECT().GetByToken(gomock.Any(), "error_token").Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rec := httptest.NewRecorder()

			handler := authMiddleware.Authenticate(nextHandler)
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.checkContext {
				assert.Equal(t, "authenticated", rec.Body.String())
			}
		})
	}
}
