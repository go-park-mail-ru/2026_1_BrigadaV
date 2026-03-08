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
	ID           uint64
	Login        string
	PasswordHash string
	FullName     string
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
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	SessionToken string `json:"session_token"`
	UserID       uint64 `json:"user_id"`
	Login        string `json:"login"`
	FullName     string `json:"full_name"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

var (
	users        = make(map[uint64]User)
	usersByLogin = make(map[string]uint64)
	sessions     = make(map[string]Session)
	places       []Place
	nextUserID   uint64 = 1
	mu           sync.RWMutex
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
	var algorithm string
	var version int
	var m, t, p int
	var saltBase64, hashBase64 string

	n, err := fmt.Sscanf(encodedHash, "%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		&algorithm, &version, &m, &t, &p, &saltBase64, &hashBase64)
	if err != nil || n != 7 {
		return false, fmt.Errorf("invalid hash format")
	}
	if algorithm != "argon2id" {
		return false, fmt.Errorf("unsupported algorithm")
	}
	salt, err := base64.RawStdEncoding.DecodeString(saltBase64)
	if err != nil {
		return false, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(hashBase64)
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

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "UNAUTHORIZED", Message: "Missing token"})
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		mu.RLock()
		session, exists := sessions[token]
		mu.RUnlock()
		if !exists || time.Now().After(session.ExpiresAt) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "UNAUTHORIZED", Message: "Invalid or expired token"})
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
		return
	}
	defer r.Body.Close()

	mu.RLock()
	userID, ok := usersByLogin[req.Login]
	var user User
	if ok {
		user = users[userID]
	}
	mu.RUnlock()

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

	mu.Lock()
	sessions[token] = session
	mu.Unlock()

	json.NewEncoder(w).Encode(LoginResponse{
		SessionToken: token,
		UserID:       user.ID,
		Login:        user.Login,
		FullName:     user.FullName,
	})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "METHOD_NOT_ALLOWED", Message: "Use POST"})
		return
	}
	token, ok := r.Context().Value("session_token").(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "UNAUTHORIZED", Message: "No session"})
		return
	}
	mu.Lock()
	delete(sessions, token)
	mu.Unlock()
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

	places = append(places,
		Place{
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
		},
		Place{
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
		},
		Place{
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
		},
	)
}

func main() {
	initPlaces()

	hashed, _ := hashPassword("123456")
	john := User{
		ID:           nextUserID,
		Login:        "BMSTU",
		PasswordHash: hashed,
		FullName:     "Daunka",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	users[john.ID] = john
	usersByLogin["BMSTU"] = john.ID
	nextUserID++

	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/logout", authMiddleware(logoutHandler))
	http.HandleFunc("/api/places", authMiddleware(placesHandler))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
