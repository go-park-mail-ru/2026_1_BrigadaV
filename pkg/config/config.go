package config

import (
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	FrontendURL        string
	AllowedOrigins     []string
	SecureCookies      bool
	YandexClientID     string
	YandexClientSecret string
	YandexRedirectURL  string
}

func Load() (*Config, error) {
	godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL is required but not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET is required but not set")
	}
	if len(jwtSecret) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 characters long")
	}

	frontendURL := getEnv("FRONTEND_URL", "http://localhost:5173")

	rawOrigins := getEnv("ALLOWED_ORIGINS", frontendURL)
	var origins []string
	for _, o := range strings.Split(rawOrigins, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	return &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        dbURL,
		JWTSecret:          jwtSecret,
		FrontendURL:        frontendURL,
		AllowedOrigins:     origins,
		SecureCookies:      getEnvBool("SECURE_COOKIES", true),
		YandexClientID:     getEnv("YANDEX_CLIENT_ID", ""),
		YandexClientSecret: getEnv("YANDEX_CLIENT_SECRET", ""),
		YandexRedirectURL:  getEnv("YANDEX_REDIRECT_URL", "http://localhost:8080/api/auth/yandex/callback"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val == "true" || val == "1" || val == "yes"
}
