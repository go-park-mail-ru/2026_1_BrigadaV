package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"guidely-app/internal/auth/repository/mocks"
	"guidely-app/pkg/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestYandexOAuthHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	handler := NewYandexOAuthHandler(
		"clientID", "secret", "http://localhost/callback", "http://frontend.com",
		false, mockUserRepo, mockSessionRepo,
	)

	req := httptest.NewRequest("GET", "/auth/yandex/login", nil)
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Contains(t, resp["url"], "oauth.yandex.ru/authorize")
	assert.Contains(t, resp["url"], "client_id=clientID")
	assert.Contains(t, resp["url"], "redirect_uri=http%3A%2F%2Flocalhost%2Fcallback")

	cookies := w.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "oauth_state" {
			found = true
			assert.NotEmpty(t, c.Value)
			break
		}
	}
	assert.True(t, found)
}

func TestYandexOAuthHandler_Callback_InvalidState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	handler := NewYandexOAuthHandler(
		"clientID", "secret", "http://localhost/callback", "http://frontend.com",
		false, mockUserRepo, mockSessionRepo,
	)

	req := httptest.NewRequest("GET", "/auth/yandex/callback?code=123&state=wrong", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "expected"})
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestYandexOAuthHandler_Callback_MissingCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	handler := NewYandexOAuthHandler(
		"clientID", "secret", "http://localhost/callback", "http://frontend.com",
		false, mockUserRepo, mockSessionRepo,
	)

	req := httptest.NewRequest("GET", "/auth/yandex/callback?state=ok", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "ok"})
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestYandexOAuthHandler_findOrCreate_ExistingByYandexID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	handler := NewYandexOAuthHandler(
		"clientID", "secret", "http://localhost/callback", "http://frontend.com",
		false, mockUserRepo, mockSessionRepo,
	)

	existingUser := &models.User{ID: 100, Nickname: "testuser"}
	yi := &yandexUserInfo{ID: "12345", Login: "testuser", DefaultEmail: "test@yandex.ru"}

	mockUserRepo.EXPECT().GetByYandexID(gomock.Any(), "12345").Return(existingUser, nil)

	user, err := handler.findOrCreate(context.Background(), yi)
	assert.NoError(t, err)
	assert.Equal(t, existingUser, user)
}

func TestYandexOAuthHandler_findOrCreate_NewUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	handler := NewYandexOAuthHandler(
		"clientID", "secret", "http://localhost/callback", "http://frontend.com",
		false, mockUserRepo, mockSessionRepo,
	)

	yi := &yandexUserInfo{ID: "12345", Login: "testuser", DefaultEmail: "test@yandex.ru", AvatarID: "avatar123"}

	mockUserRepo.EXPECT().GetByYandexID(gomock.Any(), "12345").Return(nil, nil)
	mockUserRepo.EXPECT().GetByNickname(gomock.Any(), "testuser").Return(nil, nil)
	mockUserRepo.EXPECT().GetByLogin(gomock.Any(), "test@yandex.ru").Return(nil, nil)
	mockUserRepo.EXPECT().CreateOAuth(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, user *models.User) error {
		user.ID = 200
		return nil
	})

	user, err := handler.findOrCreate(context.Background(), yi)
	assert.NoError(t, err)
	assert.Equal(t, uint64(200), user.ID)
	assert.Equal(t, "testuser", user.Nickname)
	assert.Equal(t, "test@yandex.ru", user.Login)
	assert.Equal(t, "https://avatars.yandex.net/get-yapic/avatar123/islands-200", user.AvatarURL)
	assert.Equal(t, "12345", *user.YandexID)
}

func TestYandexOAuthHandler_findOrCreate_NicknameConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	handler := NewYandexOAuthHandler(
		"clientID", "secret", "http://localhost/callback", "http://frontend.com",
		false, mockUserRepo, mockSessionRepo,
	)

	yi := &yandexUserInfo{ID: "12345", Login: "testuser", DefaultEmail: "test@yandex.ru", AvatarID: ""}

	mockUserRepo.EXPECT().GetByYandexID(gomock.Any(), "12345").Return(nil, nil)
	// Никнейм уже занят
	mockUserRepo.EXPECT().GetByNickname(gomock.Any(), "testuser").Return(&models.User{ID: 1}, nil)
	// Логин не занят
	mockUserRepo.EXPECT().GetByLogin(gomock.Any(), "test@yandex.ru").Return(nil, nil)
	mockUserRepo.EXPECT().CreateOAuth(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, user *models.User) error {
		user.ID = 200
		return nil
	})

	user, err := handler.findOrCreate(context.Background(), yi)
	assert.NoError(t, err)
	assert.Equal(t, uint64(200), user.ID)
	assert.Contains(t, user.Nickname, "testuser_") // проверяем, что добавлен суффикс
}

func TestYandexOAuthHandler_findOrCreate_LoginConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)

	handler := NewYandexOAuthHandler(
		"clientID", "secret", "http://localhost/callback", "http://frontend.com",
		false, mockUserRepo, mockSessionRepo,
	)

	yi := &yandexUserInfo{ID: "12345", Login: "testuser", DefaultEmail: "test@yandex.ru"}

	mockUserRepo.EXPECT().GetByYandexID(gomock.Any(), "12345").Return(nil, nil)
	mockUserRepo.EXPECT().GetByNickname(gomock.Any(), "testuser").Return(nil, nil)
	// Логин занят
	mockUserRepo.EXPECT().GetByLogin(gomock.Any(), "test@yandex.ru").Return(&models.User{ID: 2}, nil)
	mockUserRepo.EXPECT().CreateOAuth(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, user *models.User) error {
		user.ID = 200
		return nil
	})

	user, err := handler.findOrCreate(context.Background(), yi)
	assert.NoError(t, err)
	assert.Equal(t, uint64(200), user.ID)
	assert.Contains(t, user.Login, "+") // проверяем, что логин изменён
}

func TestGenerateOAuthState(t *testing.T) {
	state1 := generateOAuthState()
	state2 := generateOAuthState()
	assert.NotEmpty(t, state1)
	assert.Len(t, state1, 32) // 16 байт -> 32 hex символа
	assert.NotEqual(t, state1, state2)
}

// Добавить в конец файла
func TestYandexOAuthHandler_exchangeCode_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	handler := NewYandexOAuthHandler("cid", "secret", "http://cb", "http://front", false, mockUserRepo, mockSessionRepo)
	_, err := handler.exchangeCode("invalid")
	assert.Error(t, err)
}
