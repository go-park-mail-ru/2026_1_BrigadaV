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