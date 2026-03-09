// Package main GUIDELY API
//
// Документация для API сервиса Guidely
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

type User struct {
	ID           uint64    `json:"id"`
	Email        string    `json:"email"`
	Nickname     string    `json:"nickname,omitempty"`
	FullName     string    `json:"full_name,omitempty"`
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
	ID          uint64
	Name        string
	Description string
}

type Locality struct {
	ID        uint64
	Name      string
	Country   string
	Latitude  float64
	Longitude float64
}

type PlacePhoto struct {
	ID       uint64
	PlaceID  uint64
	FilePath string
	IsMain   bool
}

type Place struct {
	ID          uint64
	Name        string
	Description string
	Locality    Locality
	Category    Category
	Photos      []PlacePhoto
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PlaceResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Locality    struct {
		Name      string  `json:"name"`
		Country   string  `json:"country"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"locality"`
	Category *struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"category,omitempty"`
	Photos []struct {
		FilePath string `json:"file_path"`
		IsMain   bool   `json:"is_main"`
	} `json:"photos,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UserID   uint64 `json:"user_id"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type RegisterResponse struct {
	ID        uint64    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
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
	nextUserID      = uint64(1)
	mu              sync.RWMutex
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

func placesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "METHOD_NOT_ALLOWED", Message: "Use GET"})
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	response := make([]PlaceResponse, 0, len(places))
	for _, p := range places {
		pr := PlaceResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Locality: struct {
				Name      string  `json:"name"`
				Country   string  `json:"country"`
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			}{
				Name:      p.Locality.Name,
				Country:   p.Locality.Country,
				Latitude:  p.Locality.Latitude,
				Longitude: p.Locality.Longitude,
			},
			CreatedAt: p.CreatedAt,
		}
		if p.Category.ID != 0 {
			pr.Category = &struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			}{
				Name:        p.Category.Name,
				Description: p.Category.Description,
			}
		}
		if len(p.Photos) > 0 {
			pr.Photos = make([]struct {
				FilePath string `json:"file_path"`
				IsMain   bool   `json:"is_main"`
			}, len(p.Photos))
			for i, ph := range p.Photos {
				pr.Photos[i] = struct {
					FilePath string `json:"file_path"`
					IsMain   bool   `json:"is_main"`
				}{
					FilePath: ph.FilePath,
					IsMain:   ph.IsMain,
				}
			}
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
		Locality:    locNY,
		Category:    catPark,
		Photos: []PlacePhoto{
			{ID: 3, PlaceID: 3, FilePath: "/photos/statue.jpg", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (h *Handlers) validateRegisterRequest(req RegisterRequest) *ErrorResponse {
	log.Printf("Валидация запроса для email: %s", req.Email)

	if req.Email == "" {
		log.Println("Ошибка валидации: пустой email")
		return &ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Field:   "email",
			Message: "Email не может быть пустым",
		}
	}

	if !contains(req.Email, "@") {
		log.Printf("Ошибка валидации: email без @ - %s", req.Email)
		return &ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Field:   "email",
			Message: "Введите корректный email",
		}
	}

	if len(req.Password) < 8 {
		log.Println("Ошибка валидации: пароль короче 8 символов")
		return &ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Field:   "password",
			Message: "Пароль должен содержать не менее 8 символов",
		}
	}

	if req.FullName == "" {
		log.Println("Ошибка валидации: пустое имя")
		return &ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Field:   "full_name",
			Message: "Имя не может быть пустым",
		}
	}

	log.Printf("Валидация успешно пройдена для email: %s", req.Email)
	return nil
}

type Handlers struct {
	users        map[uint64]User
	usersByEmail map[string]uint64
	nextID       uint64
	mu           *sync.Mutex
}

func (h *Handlers) findUserByEmail(email string) (User, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if userID, exists := h.usersByEmail[email]; exists {
		return h.users[userID], true
	}
	return User{}, false
}

func (h *Handlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	log.Printf("Получен запрос на регистрацию с IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		log.Printf("Ошибка: неверный метод %s от %s", r.Method, r.RemoteAddr)
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "METHOD_NOT_ALLOWED",
			Message: "Метод не поддерживается. Используйте POST",
		})
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Ошибка декодирования JSON от %s: %v", r.RemoteAddr, err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Неверный формат запроса",
		})
		return
	}
	defer r.Body.Close()

	if errResp := h.validateRegisterRequest(req); errResp != nil {
		log.Printf("Ошибка валидации для %s: %s", req.Email, errResp.Message)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.findUserByEmail(req.Email); exists {
		log.Printf("Конфликт: email %s уже зарегистрирован", req.Email)
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "EMAIL_ALREADY_EXISTS",
			Field:   "email",
			Message: "Email уже существует",
		})
		return
	}
	log.Printf("Email %s свободен", req.Email)

	hash, salt, err := hashPasswordForRegister(req.Password)
	if err != nil {
		log.Printf("Ошибка хеширования пароля для %s: %v", req.Email, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Внутренняя ошибка сервера",
		})
		return
	}
	log.Println("Пароль успешно захеширован")

	passwordHash := encodeHash(salt, hash)
	now := time.Now()

	user := User{
		ID:           h.nextID,
		Email:        req.Email,
		PasswordHash: passwordHash,
		FullName:     req.FullName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	h.users[user.ID] = user
	h.usersByEmail[user.Email] = user.ID
	h.nextID++

	log.Printf("Пользователь сохранен. Текущее количество пользователей: %d", len(h.users))

	response := RegisterResponse{
		ID:        user.ID,
		Email:     req.Email,
		FullName:  req.FullName,
		CreatedAt: user.CreatedAt,
		Message:   "Регистрация прошла успешно",
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
	log.Printf("Успешная регистрация: %s (ID: %d)", user.Email, user.ID)
}

func main() {
	initPlaces()

	handlers := &Handlers{
		users:        make(map[uint64]User),
		usersByEmail: make(map[string]uint64),
		nextID:       1,
		mu:           &sync.Mutex{},
	}

	hashed, _ := hashPassword("123456")
	john := User{
		ID:           nextUserID,
		Email:        "john@example.com",
		Nickname:     "johnny",
		FullName:     "John Doe",
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
	http.HandleFunc("/api/places", authenticate(placesHandler))
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}