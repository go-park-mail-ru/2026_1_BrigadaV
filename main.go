// Package main GUIDELY API
//
// # Documentation for Guidely API service
//
// Schemes: http
// Host: localhost:8080
// BasePath: /
// Version: 1.0.0
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
//
// swagger:meta
package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	_ "guidely-app/docs"

	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/crypto/argon2"
)

const (
	saltLength    = 16
	keyLength     = 32
	argon2Time    = 1
	argon2Memory  = 64 * 1024
	argon2Threads = 4
)

// User represents a user in the system
type User struct {
	ID           uint64    `json:"id"`
	Login        string    `json:"login"`
	Nickname     string    `json:"nickname"`
	AvatarURL    string    `json:"avatar_url"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Session represents a user session
type Session struct {
	Token     string
	UserID    uint64
	ExpiresAt time.Time
}

// Category represents a place category
type Category struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Locality represents a geographical location
type Locality struct {
	ID        uint64  `json:"id"`
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// PlacePhoto represents a photo of a place
type PlacePhoto struct {
	ID       uint64 `json:"id"`
	PlaceID  uint64 `json:"place_id"`
	FilePath string `json:"file_path"`
	IsMain   bool   `json:"is_main"`
}

// Place represents a tourist place
type Place struct {
	ID          uint64       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Price       int64        `json:"price"`
	Locality    Locality     `json:"locality"`
	Category    Category     `json:"category"`
	Photos      []PlacePhoto `json:"photos"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// PlaceResponse represents the response for places endpoint
// swagger:response placeResponse
type PlaceResponse struct {
	ID          uint64       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Price       int64        `json:"price"`
	IsLiked     bool         `json:"is_liked"`
	Locality    Locality     `json:"locality"`
	Category    *Category    `json:"category,omitempty"`
	Photos      []PlacePhoto `json:"photos,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
}

// LoginRequest represents the login request body
// swagger:parameters login
type LoginRequest struct {
	// User login (email)
	// required: true
	// example: john@example.com
	Login string `json:"login"`

	// User password
	// required: true
	// min length: 8
	// example: 12345678
	Password string `json:"password"`
}

// LoginResponse represents the login response
// swagger:response loginResponse
type LoginResponse struct {
	UserID    uint64 `json:"user_id"`
	Login     string `json:"login"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// RegisterRequest represents the registration request body
// swagger:parameters register
type RegisterRequest struct {
	// User login (email)
	// required: true
	// example: newuser@example.com
	Login string `json:"login"`

	// User password (min 8 characters)
	// required: true
	// min length: 8
	// example: password123
	Password string `json:"password"`

	// User nickname
	// required: true
	// example: newbie
	Nickname string `json:"nickname"`
}

// RegisterResponse represents the registration response
// swagger:response registerResponse
type RegisterResponse struct {
	ID        uint64    `json:"id"`
	Login     string    `json:"login"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message,omitempty"`
}

// ErrorResponse represents an error response
// swagger:response errorResponse
type ErrorResponse struct {
	Error   string `json:"error"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

var (
	places     = make(map[uint64]Place)
	userLikes  = make(map[uint64]map[uint64]bool)
	likesMu    sync.RWMutex
	nextUserID = uint64(1)
)

type Handlers struct {
	users           map[uint64]User
	usersByLogin    map[string]uint64
	usersByNickname map[string]uint64
	sessions        map[string]Session
	nextID          uint64
	mu              sync.RWMutex
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	return salt, err
}

func hashPassword(password string) (string, error) {
	salt, err := generateSalt()
	if err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, keyLength)
	saltBase64 := base64.RawStdEncoding.EncodeToString(salt)
	hashBase64 := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argon2Memory, argon2Time, argon2Threads, saltBase64, hashBase64), nil
}

func hashPasswordForRegister(password string) ([]byte, []byte, error) {
	salt, err := generateSalt()
	if err != nil {
		return nil, nil, err
	}
	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, keyLength)
	return hash, salt, nil
}

func checkPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 5 {
		return false, fmt.Errorf("invalid hash format")
	}
	if parts[0] != "argon2id" {
		return false, fmt.Errorf("unsupported algorithm")
	}
	var m, t, p int
	_, err := fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &m, &t, &p)
	if err != nil {
		return false, fmt.Errorf("failed to parse parameters")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	newHash := argon2.IDKey([]byte(password), salt, uint32(t), uint32(m), uint8(p), uint32(len(hash)))
	return subtle.ConstantTimeCompare(newHash, hash) == 1, nil
}

func encodeHash(salt, hash []byte) string {
	saltBase64 := base64.RawStdEncoding.EncodeToString(salt)
	hashBase64 := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argon2Memory, argon2Time, argon2Threads, saltBase64, hashBase64)
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	return base64.URLEncoding.EncodeToString(b), err
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://212.233.96.48")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (h *Handlers) authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "UNAUTHORIZED", Message: "Missing session cookie"})
			return
		}

		token := cookie.Value
		h.mu.RLock()
		session, exists := h.sessions[token]
		h.mu.RUnlock()
		if !exists || time.Now().After(session.ExpiresAt) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "UNAUTHORIZED", Message: "Invalid or expired session"})
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// loginHandler handles user login
// @Summary Login user
// @Description Authenticates user and returns session cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/login [post]
func (h *Handlers) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "METHOD_NOT_ALLOWED", Message: "Use POST"})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Login decode error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "INVALID_REQUEST", Message: "Invalid JSON"})
		return
	}
	log.Printf("Login attempt: %s", req.Login)
	defer r.Body.Close()

	h.mu.RLock()
	userID, ok := h.usersByLogin[req.Login]
	var user User
	if ok {
		user = h.users[userID]
	}
	h.mu.RUnlock()

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "INVALID_CREDENTIALS", Message: "Invalid login or password"})
		return
	}

	valid, err := checkPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "INVALID_CREDENTIALS", Message: "Invalid login or password"})
		return
	}

	token, err := generateSessionToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "SERVER_ERROR", Message: "Failed to generate token"})
		return
	}

	session := Session{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	h.mu.Lock()
	h.sessions[token] = session
	h.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	json.NewEncoder(w).Encode(LoginResponse{
		UserID:    user.ID,
		Login:     user.Login,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
	})
}

// logoutHandler handles user logout
// @Summary Logout user
// @Description Invalidates user session
// @Tags auth
// @Success 200 {object} map[string]string
// @Failure 401 {object} ErrorResponse
// @Router /api/logout [post]
func (h *Handlers) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "METHOD_NOT_ALLOWED", Message: "Use POST"})
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "UNAUTHORIZED", Message: "No session"})
		return
	}
	token := cookie.Value

	h.mu.Lock()
	delete(h.sessions, token)
	h.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out"})
}

// placesHandler returns list of places
// @Summary Get places
// @Description Returns list of tourist places
// @Tags places
// @Produce json
// @Success 200 {array} PlaceResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/ [get]
func (h *Handlers) placesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "METHOD_NOT_ALLOWED", Message: "Use GET"})
		return
	}

	userIDVal := r.Context().Value("user_id")
	var userID uint64
	if userIDVal != nil {
		if id, ok := userIDVal.(uint64); ok {
			userID = id
		}
	}

	response := make([]PlaceResponse, 0, len(places))
	for _, p := range places {
		likesMu.RLock()
		liked := false
		if userLikesMap, exists := userLikes[userID]; exists {
			_, liked = userLikesMap[p.ID]
		}
		likesMu.RUnlock()

		pr := PlaceResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			IsLiked:     liked,
			Locality:    p.Locality,
			CreatedAt:   p.CreatedAt,
		}
		if p.Category.ID != 0 {
			pr.Category = &p.Category
		}
		if len(p.Photos) > 0 {
			pr.Photos = p.Photos
		}
		response = append(response, pr)
	}

	json.NewEncoder(w).Encode(response)
}

// userMeHandler returns current user info based on session cookie
// @Summary Get current user
// @Description Returns authenticated user's data
// @Tags auth
// @Produce json
// @Success 200 {object} LoginResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/user/me [get]
func (h *Handlers) userMeHandler(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		http.Error(w, `{"error":"UNAUTHORIZED","message":"Not authenticated"}`, http.StatusUnauthorized)
		return
	}

	h.mu.RLock()
	user, exists := h.users[userID]
	h.mu.RUnlock()
	if !exists {
		http.Error(w, `{"error":"UNAUTHORIZED","message":"User not found"}`, http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(LoginResponse{
		UserID:    user.ID,
		Login:     user.Login,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
	})
}

func initPlaces() {
	catHotel := Category{ID: 1, Name: "Hotel", Description: "Hotels and accommodations"}
	catMuseum := Category{ID: 2, Name: "Museum", Description: "Art and historical museums"}
	catHistorical := Category{ID: 3, Name: "Historical Site", Description: "Ancient ruins and landmarks"}
	catSquare := Category{ID: 4, Name: "Square", Description: "Public squares and plazas"}
	catResort := Category{ID: 5, Name: "Resort", Description: "Resorts and retreats"}

	locGramado := Locality{ID: 1, Name: "Грамаду", Country: "Бразилия", Latitude: -29.3733, Longitude: -50.8762}
	locParis := Locality{ID: 2, Name: "Париж", Country: "Франция", Latitude: 48.8566, Longitude: 2.3522}
	locRome := Locality{ID: 3, Name: "Рим", Country: "Италия", Latitude: 41.9028, Longitude: 12.4964}
	locBarcelona := Locality{ID: 4, Name: "Барселона", Country: "Испания", Latitude: 41.3851, Longitude: 2.1734}
	locAmsterdam := Locality{ID: 5, Name: "Амстердам", Country: "Нидерланды", Latitude: 52.3676, Longitude: 4.9041}
	locBali := Locality{ID: 6, Name: "Бали", Country: "Индонезия", Latitude: -8.4095, Longitude: 115.1889}

	now := time.Now()

	places[1] = Place{
		ID:          1,
		Name:        "Hotel Estalagem St Hubertus",
		Description: "Charming hotel in Gramado",
		Price:       2370000,
		Locality:    locGramado,
		Category:    catHotel,
		Photos: []PlacePhoto{
			{ID: 1, PlaceID: 1, FilePath: "public/mock/place/rcmd1.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	places[2] = Place{
		ID:          2,
		Name:        "Hotel Ritta Höppner",
		Description: "Cozy hotel in Gramado",
		Price:       1138100,
		Locality:    locGramado,
		Category:    catHotel,
		Photos: []PlacePhoto{
			{ID: 2, PlaceID: 2, FilePath: "public/mock/place/rcmd2.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	places[3] = Place{
		ID:          3,
		Name:        "Rodin Musée",
		Description: "Museum dedicated to Auguste Rodin",
		Price:       126900,
		Locality:    locParis,
		Category:    catMuseum,
		Photos: []PlacePhoto{
			{ID: 3, PlaceID: 3, FilePath: "public/mock/place/rcmd3.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	places[4] = Place{
		ID:          4,
		Name:        "Roman Forum",
		Description: "Ancient Roman forum",
		Price:       126900,
		Locality:    locRome,
		Category:    catHistorical,
		Photos: []PlacePhoto{
			{ID: 4, PlaceID: 4, FilePath: "public/mock/place/rcmd4.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	places[5] = Place{
		ID:          5,
		Name:        "Basílica de Santa María del Pi",
		Description: "Gothic church in Barcelona",
		Price:       199400,
		Locality:    locBarcelona,
		Category:    catHistorical,
		Photos: []PlacePhoto{
			{ID: 5, PlaceID: 5, FilePath: "public/mock/place/rcmd5.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	places[6] = Place{
		ID:          6,
		Name:        "De Hallen Amsterdam",
		Description: "Cultural complex in Amsterdam",
		Price:       3398800,
		Locality:    locAmsterdam,
		Category:    catMuseum,
		Photos: []PlacePhoto{
			{ID: 6, PlaceID: 6, FilePath: "public/mock/place/rcmd6.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	places[7] = Place{
		ID:          7,
		Name:        "Amnaya Resort Kuta",
		Description: "Resort in Bali",
		Price:       584400,
		Locality:    locBali,
		Category:    catResort,
		Photos: []PlacePhoto{
			{ID: 7, PlaceID: 7, FilePath: "public/mock/place/rcmd7.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	places[8] = Place{
		ID:          8,
		Name:        "Plaça Reial",
		Description: "Historic square in Barcelona",
		Price:       1236900,
		Locality:    locBarcelona,
		Category:    catSquare,
		Photos: []PlacePhoto{
			{ID: 8, PlaceID: 8, FilePath: "public/mock/place/rcmd8.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// HandleRegister handles user registration
// @Summary Register new user
// @Description Creates a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration data"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/register [post]
func (h *Handlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	log.Println("1. Начало HandleRegister")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "METHOD_NOT_ALLOWED", Message: "Use POST"})
		return
	}

	log.Printf("Регистрация: %s", r.RemoteAddr)

	log.Println("2. Декодирование запроса")
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("3. Ошибка декодирования: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "INVALID_REQUEST", Message: "Invalid JSON"})
		return
	}
	defer r.Body.Close()
	log.Printf("3. Запрос декодирован: %+v", req)

	log.Println("4. Валидация запроса")
	if errResp := h.validateRegisterRequest(req); errResp != nil {
		log.Printf("5. Ошибка валидации: %+v", errResp)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp)
		return
	}
	log.Println("5. Валидация пройдена")

	log.Println("6. Блокировка мьютекса")
	h.mu.Lock()
	defer h.mu.Unlock()
	log.Println("7. Мьютекс захвачен")

	log.Println("8. Проверка существования логина")
	if _, exists := h.usersByLogin[req.Login]; exists {
		log.Printf("9. Логин уже существует: %s", req.Login)
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "LOGIN_ALREADY_EXISTS",
			Field:   "login",
			Message: "Логин уже существует",
		})
		return
	}
	log.Println("9. Логин свободен")

	log.Println("10. Проверка никнейма")
	_, nicknameExists := h.usersByNickname[req.Nickname]
	if nicknameExists {
		log.Printf("11. Никнейм уже занят: %s", req.Nickname)
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "NICKNAME_ALREADY_EXISTS",
			Field:   "nickname",
			Message: "Никнейм уже занят",
		})
		return
	}
	log.Println("11. Никнейм свободен")

	log.Println("12. Хеширование пароля")
	hash, salt, err := hashPasswordForRegister(req.Password)
	if err != nil {
		log.Printf("13. Ошибка хеширования: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Внутренняя ошибка сервера",
		})
		return
	}
	log.Println("13. Пароль захеширован")

	log.Println("14. Кодирование хеша")
	passwordHash := encodeHash(salt, hash)
	now := time.Now()
	log.Println("15. Хеш закодирован")

	log.Println("16. Создание пользователя")
	user := User{
		ID:           h.nextID,
		Login:        req.Login,
		Nickname:     req.Nickname,
		AvatarURL:    "",
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	log.Printf("17. Пользователь создан: %+v", user)

	log.Println("18. Сохранение в мапы")
	h.users[user.ID] = user
	h.usersByLogin[user.Login] = user.ID
	h.usersByNickname[user.Nickname] = user.ID
	h.nextID++
	log.Println("19. Пользователь сохранен")

	log.Println("20. Инициализация лайков")
	likesMu.Lock()
	userLikes[user.ID] = make(map[uint64]bool)
	likesMu.Unlock()
	log.Println("21. Лайки инициализированы")

	log.Println("22. Формирование ответа")
	response := RegisterResponse{
		ID:        user.ID,
		Login:     req.Login,
		Nickname:  req.Nickname,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt,
		Message:   "Регистрация прошла успешно",
	}

	log.Println("23. Отправка ответа")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
	log.Printf("24. Успешная регистрация: %s (%s)", user.Login, user.Nickname)
}

func (h *Handlers) validateRegisterRequest(req RegisterRequest) *ErrorResponse {
	if req.Login == "" {
		return &ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Field:   "login",
			Message: "Логин не может быть пустым",
		}
	}

	if !contains(req.Login, "@") {
		return &ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Field:   "login",
			Message: "Введите корректный email",
		}
	}

	if len(req.Password) < 8 {
		return &ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Field:   "password",
			Message: "Пароль должен содержать не менее 8 символов",
		}
	}

	if req.Nickname == "" {
		return &ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Field:   "nickname",
			Message: "Никнейм не может быть пустым",
		}
	}

	return nil
}

func main() {
	initPlaces()

	likesMu.Lock()
	userLikes[1] = make(map[uint64]bool)
	likesMu.Unlock()

	handlers := &Handlers{
		users:           make(map[uint64]User),
		usersByLogin:    make(map[string]uint64),
		usersByNickname: make(map[string]uint64),
		sessions:        make(map[string]Session),
		nextID:          1,
	}

	hashed, _ := hashPassword("123456")
	john := User{
		ID:           nextUserID,
		Login:        "john",
		Nickname:     "johnny",
		AvatarURL:    "public/mock/user-avatar/john.jpg",
		PasswordHash: hashed,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	handlers.users[john.ID] = john
	handlers.usersByLogin[john.Login] = john.ID
	handlers.usersByNickname[john.Nickname] = john.ID
	nextUserID++

	http.HandleFunc("/api/register", corsMiddleware(handlers.HandleRegister))
	http.HandleFunc("/api/login", corsMiddleware(handlers.loginHandler))
	http.HandleFunc("/api/logout", corsMiddleware(handlers.authenticate(handlers.logoutHandler)))
	http.HandleFunc("/api/", corsMiddleware(handlers.placesHandler))
	http.HandleFunc("/api/user/me", corsMiddleware(handlers.authenticate(handlers.userMeHandler)))
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
