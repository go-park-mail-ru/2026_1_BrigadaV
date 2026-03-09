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

type User struct {
	ID           uint64    `json:"id"`
	Email        string    `json:"email"`
	Nickname     string    `json:"nickname"`
	AvatarURL    string    `json:"avatar_url"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Session struct {
	Token     string
	UserID    uint64
	ExpiresAt time.Time
}

type Category struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Locality struct {
	ID        uint64  `json:"id"`
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type PlacePhoto struct {
	ID       uint64 `json:"id"`
	PlaceID  uint64 `json:"place_id"`
	FilePath string `json:"file_path"`
	IsMain   bool   `json:"is_main"`
}

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

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UserID    uint64 `json:"user_id"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type RegisterResponse struct {
	ID        uint64    `json:"id"`
	Email     string    `json:"email"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message,omitempty"`
}

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
	userLikes       = make(map[uint64]map[uint64]bool)
	nextUserID      = uint64(1)
	mu              sync.RWMutex
	likesMu         sync.RWMutex
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
		w.Header().Set("Access-Control-Allow-Origin", "http://212.233.96.48:80")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

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
	w.Header().Set("Access-Control-Allow-Origin", "http://212.233.96.48:80")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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
		UserID:    user.ID,
		Email:     user.Email,
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
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://212.233.96.48:80")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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
	w.Header().Set("Access-Control-Allow-Origin", "http://212.233.96.48:80")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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

	mu.RLock()
	defer mu.RUnlock()

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

func initPlaces() {
	catHotel := Category{ID: 1, Name: "Hotel", Description: "Hotels and accommodations"}
	catMuseum := Category{ID: 2, Name: "Museum", Description: "Art and historical museums"}
	catHistorical := Category{ID: 3, Name: "Historical Site", Description: "Ancient ruins and landmarks"}
	catSquare := Category{ID: 4, Name: "Square", Description: "Public squares and plazas"}
	catResort := Category{ID: 5, Name: "Resort", Description: "Resorts and retreats"}

	locGramado := Locality{ID: 1, Name: "Gramado", Country: "Brazil", Latitude: -29.3733, Longitude: -50.8762}
	locParis := Locality{ID: 2, Name: "Paris", Country: "France", Latitude: 48.8566, Longitude: 2.3522}
	locRome := Locality{ID: 3, Name: "Rome", Country: "Italy", Latitude: 41.9028, Longitude: 12.4964}
	locBarcelona := Locality{ID: 4, Name: "Barcelona", Country: "Spain", Latitude: 41.3851, Longitude: 2.1734}
	locAmsterdam := Locality{ID: 5, Name: "Amsterdam", Country: "Netherlands", Latitude: 52.3676, Longitude: 4.9041}
	locBali := Locality{ID: 6, Name: "Bali", Country: "Indonesia", Latitude: -8.4095, Longitude: 115.1889}

	now := time.Now()

	places[1] = Place{
		ID:          1,
		Name:        "Hotel Estalagem St Hubertus",
		Description: "Charming hotel in Gramado",
		Price:       2370000,
		Locality:    locGramado,
		Category:    catHotel,
		Photos: []PlacePhoto{
			{ID: 1, PlaceID: 1, FilePath: "/photos/hotel_estalagem.jpg", IsMain: true},
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
			{ID: 2, PlaceID: 2, FilePath: "/photos/hotel_ritta.jpg", IsMain: true},
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
			{ID: 3, PlaceID: 3, FilePath: "/photos/rodin.jpg", IsMain: true},
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
			{ID: 4, PlaceID: 4, FilePath: "/photos/roman_forum.jpg", IsMain: true},
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
			{ID: 5, PlaceID: 5, FilePath: "/photos/basilica_pi.jpg", IsMain: true},
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
			{ID: 6, PlaceID: 6, FilePath: "/photos/de_hallen.jpg", IsMain: true},
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
			{ID: 7, PlaceID: 7, FilePath: "/photos/amnaya.jpg", IsMain: true},
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
			{ID: 8, PlaceID: 8, FilePath: "/photos/placa_reial.jpg", IsMain: true},
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
	w.Header().Set("Access-Control-Allow-Origin", "http://212.233.96.48:80")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("Регистрация: %s", r.RemoteAddr)

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
		AvatarURL:    "",
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	h.users[user.ID] = user
	h.usersByEmail[user.Email] = user.ID
	mu.Lock()
	usersByNickname[user.Nickname] = user.ID
	mu.Unlock()

	likesMu.Lock()
	userLikes[user.ID] = make(map[uint64]bool)
	likesMu.Unlock()

	h.nextID++

	response := RegisterResponse{
		ID:        user.ID,
		Email:     req.Email,
		Nickname:  req.Nickname,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt,
		Message:   "Регистрация прошла успешно",
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
	log.Printf("Успешная регистрация: %s (%s)", user.Email, user.Nickname)
}

func main() {
	initPlaces()

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
		AvatarURL:    "/avatars/john.jpg",
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
	http.HandleFunc("/api/", placesHandler)
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
