package config

import (
	"errors"
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
	CSRFSecret     string
	S3Endpoint     string
	S3AccessKey    string
	S3SecretKey    string
	S3Bucket       string
	S3UseSSL       bool
	SecureCookies  bool
	S3Enabled      bool // новое поле
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

	frontendURL := getEnv("FRONTEND_URL", "http://localhost:3000")

	rawOrigins := getEnv("ALLOWED_ORIGINS", frontendURL)
	var origins []string
	for _, o := range strings.Split(rawOrigins, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	s3Enabled := getEnvBool("S3_ENABLED", true)

	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:1111@localhost:5432/texnopark?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		FrontendURL:    frontendURL,
		AllowedOrigins: origins,
		CSRFSecret:     getEnv("CSRF_SECRET", "32-byte-long-secret-key-here!!"),
		S3Endpoint:     getEnv("S3_ENDPOINT", "localhost:9000"),
		S3AccessKey:    getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey:    getEnv("S3_SECRET_KEY", "minioadmin"),
		S3Bucket:       getEnv("S3_BUCKET", "guidely"),
		S3UseSSL:       getEnv("S3_USE_SSL", "false") == "true",
		S3Enabled:      s3Enabled,
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
