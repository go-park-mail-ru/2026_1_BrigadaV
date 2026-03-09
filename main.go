<<<<<<< feature/backend
package main

import (
	"context"
=======
// Package main GUIDELY API
//
// Документация для API регистрации сервиса Guidely
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
>>>>>>> dev
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

<<<<<<< feature/backend
=======
	_ "guidely-app/docs"
	"github.com/swaggo/http-swagger"
>>>>>>> dev
	"golang.org/x/crypto/argon2"
)

const (
	saltLength    = 16
	keyLength     = 32
<<<<<<< feature/backend
	argon2Time    = 1
	argon2Memory  = 64 * 1024
	argon2Threads = 4
)

type User struct {
	ID           uint64
	Email        string
	Nickname     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
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

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

var (
	users                  = make(map[uint64]User)
	usersByEmail           = make(map[string]uint64)
	usersByNickname        = make(map[string]uint64)
	sessions               = make(map[string]Session)
	places                 = make(map[uint64]Place)
	nextUserID      uint64 = 1
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

func checkPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 5 {
		return false, fmt.Errorf("invalid hash format: expected 5 parts, got %d", len(parts))
	}
	if parts[0] != "argon2id" {
		return false, fmt.Errorf("unsupported algorithm: %s", parts[0])
	}
	var m, t, p int
	_, err := fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &m, &t, &p)
	if err != nil {
		return false, fmt.Errorf("failed to parse parameters: %v", err)
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

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "UNAUTHORIZED", Message: "Missing session cookie"})
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "BAD_REQUEST", Message: "Invalid cookie"})
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

		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", session.UserID)
		ctx = context.WithValue(ctx, "session_token", token)
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
=======
	argon2time    = 1
	argon2memory  = 64 << 10
	argon2threads = 4
)

func generateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func hashPassword(password string) ([]byte, []byte, error) {
	log.Println("Начало хеширования пароля")
	salt, err := generateSalt(saltLength)
	if err != nil {
		log.Printf("Ошибка генерации соли: %v", err)
		return nil, nil, err
	}
	hashed := argon2.IDKey([]byte(password), salt, argon2time, argon2memory, argon2threads, keyLength)
	log.Println("Пароль успешно захеширован")
	return hashed, salt, nil
}

func verifyPassword(password string, salt []byte, expectedHash []byte) bool {
	newHash := argon2.IDKey([]byte(password), salt, argon2time, argon2memory, argon2threads, keyLength)
	return subtleCompare(newHash, expectedHash)
}

func subtleCompare(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

func encodeHash(salt, hash []byte) string {
	saltBase64 := base64.RawStdEncoding.EncodeToString(salt)
	hashBase64 := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argon2memory, argon2time, argon2threads, saltBase64, hashBase64)
}

func decodeHash(encodedHash string) (salt, hash []byte, err error) {
	var algorithm string
	var version int
	var m, t, p int
	var saltBase64, hashBase64 string

	n, err := fmt.Sscanf(encodedHash, "%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		&algorithm, &version, &m, &t, &p, &saltBase64, &hashBase64)
	if err != nil || n != 7 {
		return nil, nil, fmt.Errorf("Неверный формат хеша")
	}

	salt, err = base64.RawStdEncoding.DecodeString(saltBase64)
	if err != nil {
		return nil, nil, err
	}
	hash, err = base64.RawStdEncoding.DecodeString(hashBase64)
	if err != nil {
		return nil, nil, err
	}
	return salt, hash, nil
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// User представляет модель пользователя в системе
type User struct {
	ID           uint64    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RegisterRequest представляет тело запроса для регистрации нового пользователя
// swagger:parameters registerUser
type RegisterRequest struct {
	// Email пользователя, используется для входа
	// required: true
	// example: user@example.com
	// pattern: ^\S+@\S+\.\S+$
	Email string `json:"email"`

	// Пароль пользователя (минимум 8 символов)
	// required: true
	// min length: 8
	// example: securePass123
	Password string `json:"password"`

	// Полное имя пользователя
	// required: true
	// example: Иван Петров
	FullName string `json:"full_name"`
}

// RegisterResponse представляет успешный ответ на регистрацию
// swagger:response registerResponse
type RegisterResponse struct {
	// Уникальный идентификатор пользователя
	// example: 1
	ID uint64 `json:"id"`

	// Email пользователя
	// example: user@example.com
	Email string `json:"email"`

	// Полное имя пользователя
	// example: Иван Петров
	FullName string `json:"full_name"`

	// Дата и время создания аккаунта
	// example: 2024-01-15T10:30:00Z
	CreatedAt time.Time `json:"created_at"`

	// Сообщение о результате операции
	// example: Регистрация прошла успешно
	Message string `json:"message,omitempty"`
}

// ErrorResponse представляет структуру ошибки
// swagger:response errorResponse
type ErrorResponse struct {
	// Код ошибки
	// example: VALIDATION_ERROR
	Error string `json:"error"`

	// Поле, в котором произошла ошибка (если применимо)
	// example: email
	Field string `json:"field,omitempty"`

	// Описание ошибки
	// example: Email не может быть пустым
	Message string `json:"message"`
}

// Handlers содержит состояние сервера и обработчики HTTP запросов
type Handlers struct {
	users  map[uint64]User
	usersByEmail map[string]uint64
	nextID uint64
	mu     *sync.Mutex
}

func (h *Handlers) findUserByEmail(email string) (User, bool) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    if userID, exists := h.usersByEmail[email]; exists {
        return h.users[userID], true
    }
    return User{}, false
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

// HandleRegister обрабатывает запросы на регистрацию новых пользователей
//
// @Summary Регистрация нового пользователя
// @Description Создаёт новую учётную запись пользователя с указанными email, паролем и именем
// @Tags users
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Данные для регистрации"
// @Success 201 {object} RegisterResponse "Пользователь успешно создан"
// @Failure 400 {object} ErrorResponse "Ошибка валидации или неверный формат запроса"
// @Failure 409 {object} ErrorResponse "Email уже зарегистрирован"
// @Failure 405 {object} ErrorResponse "Метод не поддерживается"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/register [post]
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
>>>>>>> dev
		return
	}
	defer r.Body.Close()

<<<<<<< feature/backend
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
	if err != nil {
		log.Printf("Password check error: %v", err)
	}
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

func main() {
	initPlaces()

	hashed, err := hashPassword("123456")
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}
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

	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/logout", authenticate(logoutHandler))
	http.HandleFunc("/api/places", authenticate(placesHandler))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
=======
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

	hash, salt, err := hashPassword(req.Password)
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
	handlers := &Handlers{
		users:  make(map[uint64]User),
		usersByEmail: make(map[string]uint64),
		nextID: 1,
		mu:     &sync.Mutex{},
	}

	http.HandleFunc("/api/register", handlers.HandleRegister)

	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	port := ":8080"
	log.Printf("Сервер запущен на http://localhost%s", port)
	log.Printf("Swagger доступен на http://localhost%s/swagger/index.html", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
>>>>>>> dev
