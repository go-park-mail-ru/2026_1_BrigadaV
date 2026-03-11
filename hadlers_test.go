package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetGlobals() {
	places = make(map[uint64]Place)
	userLikes = make(map[uint64]map[uint64]bool)
	nextUserID = 1
	initPlaces()
}

func newTestHandlers() *Handlers {
	return &Handlers{
		users:           make(map[uint64]User),
		usersByLogin:    make(map[string]uint64),
		usersByNickname: make(map[string]uint64),
		sessions:        make(map[string]Session),
		nextID:          1,
	}
}

func TestValidateRegisterRequest(t *testing.T) {
	h := newTestHandlers()

	tests := []struct {
		name      string
		req       RegisterRequest
		wantErr   bool
		wantField string
	}{
		{"valid request", RegisterRequest{Login: "user@test.com", Password: "12345678", Nickname: "john"}, false, ""},
		{"empty login", RegisterRequest{Login: "", Password: "12345678", Nickname: "john"}, true, "login"},
		{"login without @", RegisterRequest{Login: "user", Password: "12345678", Nickname: "john"}, true, "login"},
		{"password too short", RegisterRequest{Login: "user@test.com", Password: "1234567", Nickname: "john"}, true, "password"},
		{"empty nickname", RegisterRequest{Login: "user@test.com", Password: "12345678", Nickname: ""}, true, "nickname"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errResp := h.validateRegisterRequest(tt.req)
			if tt.wantErr {
				require.NotNil(t, errResp, "expected error response")
				assert.Equal(t, tt.wantField, errResp.Field)
			} else {
				assert.Nil(t, errResp)
			}
		})
	}
}

func TestHandleRegisterSuccess(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	reqBody := RegisterRequest{Login: "new@test.com", Password: "password123", Nickname: "newbie"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleRegister(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var regResp RegisterResponse
	err := json.NewDecoder(resp.Body).Decode(&regResp)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), regResp.ID)
	assert.Equal(t, "new@test.com", regResp.Login)
	assert.Equal(t, "newbie", regResp.Nickname)
	assert.Equal(t, "Регистрация прошла успешно", regResp.Message)
}

func TestHandleRegisterLoginConflict(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	req1 := RegisterRequest{Login: "same@test.com", Password: "pass1234", Nickname: "nick1"}
	body1, _ := json.Marshal(req1)
	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body1))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleRegister(w, req)
	assert.Equal(t, http.StatusCreated, w.Result().StatusCode)

	req2 := RegisterRequest{Login: "same@test.com", Password: "pass1234", Nickname: "nick2"}
	body2, _ := json.Marshal(req2)
	req = httptest.NewRequest("POST", "/api/register", bytes.NewReader(body2))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.HandleRegister(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var errResp ErrorResponse
	err := json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "login", errResp.Field)
	assert.Contains(t, errResp.Message, "Логин уже существует")
}

func TestHandleRegisterNicknameConflict(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	req1 := RegisterRequest{Login: "user1@test.com", Password: "pass1234", Nickname: "taken"}
	body1, _ := json.Marshal(req1)
	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body1))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleRegister(w, req)
	assert.Equal(t, http.StatusCreated, w.Result().StatusCode)

	req2 := RegisterRequest{Login: "user2@test.com", Password: "pass1234", Nickname: "taken"}
	body2, _ := json.Marshal(req2)
	req = httptest.NewRequest("POST", "/api/register", bytes.NewReader(body2))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.HandleRegister(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var errResp ErrorResponse
	err := json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "nickname", errResp.Field)
	assert.Contains(t, errResp.Message, "Никнейм уже занят")
}

func TestHandleRegisterInvalidJSON(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader([]byte(`{bad json}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleRegister(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestLoginHandlerSuccess(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	hashed, _ := hashPassword("pass123")
	user := User{
		ID:           h.nextID,
		Login:        "login@test.com",
		Nickname:     "loginer",
		AvatarURL:    "",
		PasswordHash: hashed,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	h.users[user.ID] = user
	h.usersByLogin[user.Login] = user.ID
	h.usersByNickname[user.Nickname] = user.ID
	h.nextID++

	loginReq := LoginRequest{Login: "login@test.com", Password: "pass123"}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.loginHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp LoginResponse
	err := json.NewDecoder(resp.Body).Decode(&loginResp)
	require.NoError(t, err)
	assert.Equal(t, user.ID, loginResp.UserID)
	assert.Equal(t, user.Login, loginResp.Login)

	cookies := resp.Header["Set-Cookie"]
	assert.NotEmpty(t, cookies, "cookie should be set")
}

func TestLoginHandlerWrongPassword(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	hashed, _ := hashPassword("correct")
	user := User{ID: h.nextID, Login: "user@test.com", PasswordHash: hashed}
	h.users[user.ID] = user
	h.usersByLogin[user.Login] = user.ID
	h.nextID++

	loginReq := LoginRequest{Login: "user@test.com", Password: "wrong"}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.loginHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLoginHandlerUserNotFound(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	loginReq := LoginRequest{Login: "nonexistent@test.com", Password: "pass"}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.loginHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLoginHandlerInvalidJSON(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader([]byte(`{bad}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.loginHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestLogoutHandlerSuccess(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	token, _ := generateSessionToken()
	session := Session{Token: token, UserID: 1, ExpiresAt: time.Now().Add(7 * 24 * time.Hour)}
	h.sessions[token] = session

	req := httptest.NewRequest("POST", "/api/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	w := httptest.NewRecorder()
	h.logoutHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotContains(t, h.sessions, token, "session should be deleted")
	cookies := resp.Header["Set-Cookie"]
	assert.NotEmpty(t, cookies, "cookie should be cleared")
}

func TestLogoutHandlerNoCookie(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	req := httptest.NewRequest("POST", "/api/logout", nil)
	w := httptest.NewRecorder()
	h.logoutHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestPlacesHandlerUnauthenticated(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	req := httptest.NewRequest("GET", "/api/", nil)
	w := httptest.NewRecorder()
	h.placesHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var placesResp []PlaceResponse
	err := json.NewDecoder(resp.Body).Decode(&placesResp)
	require.NoError(t, err)
	assert.NotEmpty(t, placesResp, "place list should not be empty")
	for _, p := range placesResp {
		assert.False(t, p.IsLiked, "place %d should not be liked by unauthenticated user", p.ID)
	}
}

func TestPlacesHandlerAuthenticatedNoLikes(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	userID := uint64(1)
	token, _ := generateSessionToken()
	h.sessions[token] = Session{Token: token, UserID: userID, ExpiresAt: time.Now().Add(7 * 24 * time.Hour)}

	req := httptest.NewRequest("GET", "/api/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	ctx := context.WithValue(req.Context(), "user_id", userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.placesHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var placesResp []PlaceResponse
	err := json.NewDecoder(resp.Body).Decode(&placesResp)
	require.NoError(t, err)
	for _, p := range placesResp {
		assert.False(t, p.IsLiked, "place %d should not be liked without any likes", p.ID)
	}
}

func TestPlacesHandlerWithLikes(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	userID := uint64(1)
	likesMu.Lock()
	userLikes[userID] = map[uint64]bool{1: true}
	likesMu.Unlock()

	token, _ := generateSessionToken()
	h.sessions[token] = Session{Token: token, UserID: userID, ExpiresAt: time.Now().Add(7 * 24 * time.Hour)}

	req := httptest.NewRequest("GET", "/api/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	ctx := context.WithValue(req.Context(), "user_id", userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.placesHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var placesResp []PlaceResponse
	err := json.NewDecoder(resp.Body).Decode(&placesResp)
	require.NoError(t, err)

	found := false
	for _, p := range placesResp {
		if p.ID == 1 {
			assert.True(t, p.IsLiked, "place 1 should be liked")
			found = true
		} else {
			assert.False(t, p.IsLiked, "place %d should not be liked", p.ID)
		}
	}
	assert.True(t, found, "place 1 not found in response")
}
