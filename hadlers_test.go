package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
		name string
		req  RegisterRequest
		want *ErrorResponse
	}{
		{
			name: "valid request",
			req:  RegisterRequest{Login: "user@test.com", Password: "12345678", Nickname: "john"},
			want: nil,
		},
		{
			name: "empty login",
			req:  RegisterRequest{Login: "", Password: "12345678", Nickname: "john"},
			want: &ErrorResponse{Field: "login", Message: "Логин не может быть пустым"},
		},
		{
			name: "login without @",
			req:  RegisterRequest{Login: "user", Password: "12345678", Nickname: "john"},
			want: &ErrorResponse{Field: "login", Message: "Введите корректный email"},
		},
		{
			name: "password too short",
			req:  RegisterRequest{Login: "user@test.com", Password: "1234567", Nickname: "john"},
			want: &ErrorResponse{Field: "password", Message: "Пароль должен содержать не менее 8 символов"},
		},
		{
			name: "empty nickname",
			req:  RegisterRequest{Login: "user@test.com", Password: "12345678", Nickname: ""},
			want: &ErrorResponse{Field: "nickname", Message: "Никнейм не может быть пустым"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := h.validateRegisterRequest(tt.req)
			if tt.want == nil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
			} else {
				if got == nil {
					t.Errorf("expected error, got nil")
					return
				}
				if got.Field != tt.want.Field {
					t.Errorf("field = %s, want %s", got.Field, tt.want.Field)
				}
				if got.Message != tt.want.Message {
					t.Errorf("message = %s, want %s", got.Message, tt.want.Message)
				}
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
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}

	var regResp RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if regResp.ID != 1 {
		t.Errorf("expected user ID 1, got %d", regResp.ID)
	}
	if regResp.Login != "new@test.com" {
		t.Errorf("login = %s, want new@test.com", regResp.Login)
	}
	if regResp.Nickname != "newbie" {
		t.Errorf("nickname = %s, want newbie", regResp.Nickname)
	}
	if regResp.Message != "Регистрация прошла успешно" {
		t.Errorf("message = %s, want correct", regResp.Message)
	}
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
	if w.Result().StatusCode != http.StatusCreated {
		t.Fatal("failed to create first user")
	}

	req2 := RegisterRequest{Login: "same@test.com", Password: "pass1234", Nickname: "nick2"}
	body2, _ := json.Marshal(req2)
	req = httptest.NewRequest("POST", "/api/register", bytes.NewReader(body2))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.HandleRegister(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 Conflict, got %d", resp.StatusCode)
	}
	var errResp ErrorResponse
	json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Field != "login" {
		t.Errorf("expected field login, got %s", errResp.Field)
	}
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
	if w.Result().StatusCode != http.StatusCreated {
		t.Fatal("failed to create first user")
	}

	req2 := RegisterRequest{Login: "user2@test.com", Password: "pass1234", Nickname: "taken"}
	body2, _ := json.Marshal(req2)
	req = httptest.NewRequest("POST", "/api/register", bytes.NewReader(body2))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.HandleRegister(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 Conflict, got %d", resp.StatusCode)
	}
	var errResp ErrorResponse
	json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Field != "nickname" {
		t.Errorf("expected field nickname, got %s", errResp.Field)
	}
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
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", resp.StatusCode)
	}
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
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatal(err)
	}
	if loginResp.UserID != user.ID {
		t.Errorf("user_id = %d, want %d", loginResp.UserID, user.ID)
	}
	if loginResp.Login != user.Login {
		t.Errorf("login = %s, want %s", loginResp.Login, user.Login)
	}
	cookies := resp.Header["Set-Cookie"]
	if len(cookies) == 0 {
		t.Error("no Set-Cookie header")
	}
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
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
	}
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
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
	}
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
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", resp.StatusCode)
	}
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
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	if _, exists := h.sessions[token]; exists {
		t.Error("session still exists after logout")
	}
	cookies := resp.Header["Set-Cookie"]
	if len(cookies) == 0 {
		t.Error("no Set-Cookie header")
	}
}

func TestLogoutHandlerNoCookie(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	req := httptest.NewRequest("POST", "/api/logout", nil)
	w := httptest.NewRecorder()
	h.logoutHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}

func TestPlacesHandlerUnauthenticated(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	req := httptest.NewRequest("GET", "/api/", nil)
	w := httptest.NewRecorder()
	h.placesHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var placesResp []PlaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&placesResp); err != nil {
		t.Fatal(err)
	}
	if len(placesResp) == 0 {
		t.Error("empty place list")
	}
	for _, p := range placesResp {
		if p.IsLiked {
			t.Errorf("place %d has is_liked=true for unauthenticated user", p.ID)
		}
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
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var placesResp []PlaceResponse
	json.NewDecoder(resp.Body).Decode(&placesResp)
	for _, p := range placesResp {
		if p.IsLiked {
			t.Errorf("place %d has is_liked=true but user has no likes", p.ID)
		}
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
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var placesResp []PlaceResponse
	json.NewDecoder(resp.Body).Decode(&placesResp)
	found := false
	for _, p := range placesResp {
		if p.ID == 1 {
			if !p.IsLiked {
				t.Error("place 1 should be liked")
			}
			found = true
		} else {
			if p.IsLiked {
				t.Errorf("place %d should not be liked", p.ID)
			}
		}
	}
	if !found {
		t.Error("place 1 not found in response")
	}
}
