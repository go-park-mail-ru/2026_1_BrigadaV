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

	"github.com/swaggo/http-swagger"
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
	Email        string    `json:"email"`
	Nickname     string    `json:"nickname"`
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
	Price       float64      `json:"price"` // Added price field
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
	Price       float64      `json:"price"`    // Added price field
	IsLiked     bool         `json:"is_liked"` // Flag indicating if current user liked this place
	Locality    Locality     `json:"locality"`
	Category    *Category    `json:"category,omitempty"`
	Photos      []PlacePhoto `json:"photos,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
}

// LoginRequest represents the login request body
// swagger:parameters login
type LoginRequest struct {
	// User email
	// required: true
	// example: john@example.com
	Email string `json:"email"`

	// User password
	// required: true
	// example: 123456
	Password string `json:"password"`
}

// LoginResponse represents the login response
// swagger:response loginResponse
type LoginResponse struct {
	UserID   uint64 `json:"user_id"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

// RegisterRequest represents the registration request body
// swagger:parameters register
type RegisterRequest struct {
	// User email
	// required: true
	// example: newuser@example.com
	Email string `json:"email"`

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
	Email     string    `json:"email"`
	Nickname  string    `json:"nickname"`
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
	users           = make(map[uint64]User)
	usersByEmail    = make(map[string]uint64)
	usersByNickname = make(map[string]uint64)
	sessions        = make(map[string]Session)
	places          = make(map[uint64]Place)
	// userLikes stores likes: userLikes[userID][placeID] = true
	userLikes  = make(map[uint64]map[uint64]bool)
	nextUserID = uint64(1)
	mu         sync.RWMutex
	likesMu    sync.RWMutex // separate mutex for likes
)

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

func authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "UNAUTHORIZED", Message: "Missing session cookie"})
			return
		}

		token := cookie.Value
		mu.RLock()
		session, exists := sessions[token]
		mu.RUnlock()
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
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "METHOD_NOT_ALLOWED", Message: "Use POST"})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "INVALID_REQUEST", Message: "Invalid JSON"})
		return
	}
	defer r.Body.Close()

	mu.RLock()
	userID, ok := usersByEmail[req.Email]
	var user User
	if ok {
		user = users[userID]
	}
	mu.RUnlock()

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "INVALID_CREDENTIALS", Message: "Invalid email or password"})
		return
	}

	valid, err := checkPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "INVALID_CREDENTIALS", Message: "Invalid email or password"})
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

	mu.Lock()
	sessions[token] = session
	mu.Unlock()

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
		UserID:   user.ID,
		Email:    user.Email,
		Nickname: user.Nickname,
	})
}

// logoutHandler handles user logout
// @Summary Logout user
// @Description Invalidates user session
// @Tags auth
// @Success 200 {object} map[string]string
// @Failure 401 {object} ErrorResponse
// @Router /api/logout [post]
func logoutHandler(w http.ResponseWriter, r *http.Request) {
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

	mu.Lock()
	delete(sessions, token)
	mu.Unlock()

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
func placesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "METHOD_NOT_ALLOWED", Message: "Use GET"})
		return
	}

	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "INTERNAL_ERROR", Message: "Invalid user context"})
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	response := make([]PlaceResponse, 0, len(places))
	for _, p := range places {
		// Check if current user liked this place
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func initPlaces() {
	catMuseum := Category{ID: 1, Name: "Museum", Description: "Art and historical museums"}
	catPark := Category{ID: 2, Name: "Park", Description: "City parks and reserves"}

	locParis := Locality{ID: 1, Name: "Paris", Country: "France", Latitude: 48.8566, Longitude: 2.3522}
	locRome := Locality{ID: 2, Name: "Rome", Country: "Italy", Latitude: 41.9028, Longitude: 12.4964}
	locNY := Locality{ID: 3, Name: "New York", Country: "USA", Latitude: 40.7128, Longitude: -74.0060}

	now := time.Now()

	places[1] = Place{
		ID:          1,
		Name:        "Eiffel Tower",
		Description: "Famous tower in Paris",
		Price:       15.0,
		Locality:    locParis,
		Category:    catPark,
		Photos: []PlacePhoto{
			{ID: 1, PlaceID: 1, FilePath: "/photos/eiffel.jpg", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	places[2] = Place{
		ID:          2,
		Name:        "Colosseum",
		Description: "Ancient amphitheater in Rome",
		Price:       12.5,
		Locality:    locRome,
		Category:    catMuseum,
		Photos: []PlacePhoto{
			{ID: 2, PlaceID: 2, FilePath: "/photos/colosseum.jpg", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	places[3] = Place{
		ID:          3,
		Name:        "Statue of Liberty",
		Description: "Gift from France to USA",
		Price:       10.0,
		Locality:    locNY,
		Category:    catPark,
		Photos: []PlacePhoto{
			{ID: 3, PlaceID: 3, FilePath: "/photos/statue.jpg", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type Handlers struct {
	users        map[uint64]User
	usersByEmail map[string]uint64
	nextID       uint64
	mu           sync.RWMutex
}

func (h *Handlers) validateRegisterRequest(req RegisterRequest) *ErrorResponse {
	if req.Email == "" {
		return &ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Field:   "email",
			Message: "Email не может быть пустым",
		}
	}

	if !contains(req.Email, "@") {
		return &ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Field:   "email",
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
	log.Printf("Регистрация: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "METHOD_NOT_ALLOWED", Message: "Use POST"})
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "INVALID_REQUEST", Message: "Invalid JSON"})
		return
	}
	defer r.Body.Close()

	if errResp := h.validateRegisterRequest(req); errResp != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.usersByEmail[req.Email]; exists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "EMAIL_ALREADY_EXISTS",
			Field:   "email",
			Message: "Email уже существует",
		})
		return
	}

	mu.RLock()
	_, nicknameExists := usersByNickname[req.Nickname]
	mu.RUnlock()
	if nicknameExists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "NICKNAME_ALREADY_EXISTS",
			Field:   "nickname",
			Message: "Никнейм уже занят",
		})
		return
	}

	hash, salt, err := hashPasswordForRegister(req.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Внутренняя ошибка сервера",
		})
		return
	}

	passwordHash := encodeHash(salt, hash)
	now := time.Now()

	user := User{
		ID:           h.nextID,
		Email:        req.Email,
		Nickname:     req.Nickname,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	h.users[user.ID] = user
	h.usersByEmail[user.Email] = user.ID
	mu.Lock()
	usersByNickname[user.Nickname] = user.ID
	mu.Unlock()

	// Initialize empty likes map for new user
	likesMu.Lock()
	userLikes[user.ID] = make(map[uint64]bool)
	likesMu.Unlock()

	h.nextID++

	response := RegisterResponse{
		ID:        user.ID,
		Email:     req.Email,
		Nickname:  req.Nickname,
		CreatedAt: user.CreatedAt,
		Message:   "Регистрация прошла успешно",
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
	log.Printf("Успешная регистрация: %s (%s)", user.Email, user.Nickname)
}

func main() {
	initPlaces()

	// Initialize likes for existing user
	likesMu.Lock()
	userLikes[1] = make(map[uint64]bool)
	likesMu.Unlock()

	handlers := &Handlers{
		users:        make(map[uint64]User),
		usersByEmail: make(map[string]uint64),
		nextID:       1,
	}

	hashed, _ := hashPassword("123456")
	john := User{
		ID:           nextUserID,
		Email:        "john@example.com",
		Nickname:     "johnny",
		PasswordHash: hashed,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	users[john.ID] = john
	usersByEmail[john.Email] = john.ID
	usersByNickname[john.Nickname] = john.ID
	nextUserID++

	http.HandleFunc("/api/register", handlers.HandleRegister)
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/logout", authenticate(logoutHandler))
	http.HandleFunc("/api/", authenticate(placesHandler))
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
