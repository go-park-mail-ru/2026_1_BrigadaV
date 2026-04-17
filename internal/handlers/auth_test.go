package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"guidely-app/internal/dto"
	"guidely-app/internal/models"
	"guidely-app/internal/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	handler := NewAuthHandler(mockAuthService)

	reqBody := dto.RegisterRequest{
		Login:    "test@example.com",
		Password: "12345678",
		Nickname: "tester",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockAuthService.EXPECT().Register(gomock.Any(), gomock.Any()).Return(&models.User{ID: 1}, "token123", nil)

	handler.Register(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "user created", resp["message"])
	assert.Equal(t, float64(1), resp["user_id"])
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	handler := NewAuthHandler(mockAuthService)

	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader([]byte(`{invalid json}`)))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	handler := NewAuthHandler(mockAuthService)

	reqBody := dto.LoginRequest{
		Login:    "test@example.com",
		Password: "12345678",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockAuthService.EXPECT().Login(gomock.Any(), gomock.Any()).Return(&models.User{ID: 1, Nickname: "tester", AvatarURL: "/avatar.jpg"}, "token123", nil)

	handler.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.LoginResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, uint64(1), resp.UserID)
	assert.Equal(t, "tester", resp.Nickname)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	handler := NewAuthHandler(mockAuthService)

	reqBody := dto.LoginRequest{
		Login:    "test@example.com",
		Password: "wrong",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockAuthService.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, "", errors.New("invalid credentials"))

	handler.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	handler := NewAuthHandler(mockAuthService)

	req := httptest.NewRequest("POST", "/api/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "token123"})
	w := httptest.NewRecorder()

	mockAuthService.EXPECT().Logout(gomock.Any(), "token123").Return(nil)

	handler.Logout(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_Me_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	handler := NewAuthHandler(mockAuthService)

	req := httptest.NewRequest("GET", "/api/user/me", nil)
	ctx := context.WithValue(req.Context(), "user_id", uint64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockAuthService.EXPECT().GetUserByID(gomock.Any(), uint64(1)).Return(&models.User{ID: 1, Nickname: "tester", AvatarURL: "/avatar.jpg"}, nil)

	handler.Me(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.LoginResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, uint64(1), resp.UserID)
	assert.Equal(t, "tester", resp.Nickname)
}

func TestAuthHandler_Me_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	handler := NewAuthHandler(mockAuthService)

	req := httptest.NewRequest("GET", "/api/user/me", nil)
	w := httptest.NewRecorder()

	handler.Me(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
