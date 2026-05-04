package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	FrontendURL    string
	AllowedOrigins []string
}

func Load() (*Config, error) {
	godotenv.Load()

	frontendURL := getEnv("FRONTEND_URL", "http://localhost:3000")

	rawOrigins := getEnv("ALLOWED_ORIGINS", frontendURL)
	var origins []string
	for _, o := range strings.Split(rawOrigins, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:1111@localhost:5432/texnopark?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		FrontendURL:    frontendURL,
		AllowedOrigins: origins,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
