package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"guidely-app/internal/auth/repository"
	"guidely-app/pkg/models"
	"guidely-app/pkg/utils"
)

type YandexOAuthHandler struct {
	clientID     string
	clientSecret string
	redirectURL  string
	frontendURL  string
	userRepo     repository.UserRepository
	sessionRepo  repository.SessionRepository
}

func NewYandexOAuthHandler(
	clientID, clientSecret, redirectURL, frontendURL string,
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
) *YandexOAuthHandler {
	return &YandexOAuthHandler{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		frontendURL:  frontendURL,
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
	}
}

func (h *YandexOAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	state := generateOAuthState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	authURL := fmt.Sprintf(
		"https://oauth.yandex.ru/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s",
		h.clientID,
		url.QueryEscape(h.redirectURL),
		state,
	)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": authURL})
}

func (h *YandexOAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", MaxAge: -1, Path: "/"})

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	accessToken, err := h.exchangeCode(code)
	if err != nil {
		http.Error(w, "failed to exchange code", http.StatusInternalServerError)
		return
	}

	yandexUser, err := h.fetchYandexUser(accessToken)
	if err != nil {
		http.Error(w, "failed to fetch user info", http.StatusInternalServerError)
		return
	}

	user, err := h.findOrCreate(r.Context(), yandexUser)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	token, err := utils.GenerateSessionToken()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session := &models.Session{
		UserID:    user.ID,
		TokenHash: utils.HashToken(token),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := h.sessionRepo.Create(r.Context(), session); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		MaxAge:   7 * 24 * 60 * 60,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	http.Redirect(w, r, h.frontendURL, http.StatusTemporaryRedirect)
}

type yandexUserInfo struct {
	ID           string `json:"id"`
	Login        string `json:"login"`
	DefaultEmail string `json:"default_email"`
	AvatarID     string `json:"default_avatar_id"`
}

func (h *YandexOAuthHandler) exchangeCode(code string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", h.clientID)
	data.Set("client_secret", h.clientSecret)

	resp, err := http.PostForm("https://oauth.yandex.ru/token", data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("yandex token error: %s", result.Error)
	}
	return result.AccessToken, nil
}

func (h *YandexOAuthHandler) fetchYandexUser(token string) (*yandexUserInfo, error) {
	req, err := http.NewRequest("GET", "https://login.yandex.ru/info?format=json", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "OAuth "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var info yandexUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (h *YandexOAuthHandler) findOrCreate(ctx context.Context, yi *yandexUserInfo) (*models.User, error) {
	user, err := h.userRepo.GetByYandexID(ctx, yi.ID)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	nickname := yi.Login
	if ex, _ := h.userRepo.GetByNickname(ctx, nickname); ex != nil {
		nickname = nickname + "_" + yi.ID[:4]
	}

	avatarURL := ""
	if yi.AvatarID != "" {
		avatarURL = fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", yi.AvatarID)
	}

	login := yi.DefaultEmail
	if login == "" {
		login = yi.Login + "@yandex.ru"
	}

	yandexID := yi.ID
	newUser := &models.User{
		Login:     login,
		Nickname:  nickname,
		AvatarURL: avatarURL,
		YandexID:  &yandexID,
	}
	if err := h.userRepo.CreateOAuth(ctx, newUser); err != nil {
		return nil, err
	}
	return newUser, nil
}

func generateOAuthState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
