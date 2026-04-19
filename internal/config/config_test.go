package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	originalPort := os.Getenv("PORT")
	originalDatabaseURL := os.Getenv("DATABASE_URL")
	originalJWTSecret := os.Getenv("JWT_SECRET")
	originalFrontendURL := os.Getenv("FRONTEND_URL")

	defer func() {
		os.Setenv("PORT", originalPort)
		os.Setenv("DATABASE_URL", originalDatabaseURL)
		os.Setenv("JWT_SECRET", originalJWTSecret)
		os.Setenv("FRONTEND_URL", originalFrontendURL)
	}()

	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("FRONTEND_URL")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "postgres://postgres:1111@212.233.96.48:5432/texnopark?sslmode=disable", cfg.DatabaseURL)
	assert.Equal(t, "your-secret-key", cfg.JWTSecret)
	assert.Equal(t, "http://guidely.ru", cfg.FrontendURL)

	os.Setenv("PORT", "3000")
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("JWT_SECRET", "custom-secret")
	os.Setenv("FRONTEND_URL", "http://localhost:3000")

	cfg, err = Load()
	assert.NoError(t, err)
	assert.Equal(t, "3000", cfg.Port)
	assert.Equal(t, "postgres://localhost/test", cfg.DatabaseURL)
	assert.Equal(t, "custom-secret", cfg.JWTSecret)
	assert.Equal(t, "http://localhost:3000", cfg.FrontendURL)
}

func TestGetEnv(t *testing.T) {
	key := "TEST_ENV_KEY"
	defaultValue := "default"

	os.Unsetenv(key)
	val := getEnv(key, defaultValue)
	assert.Equal(t, defaultValue, val)

	os.Setenv(key, "custom")
	val = getEnv(key, defaultValue)
	assert.Equal(t, "custom", val)

	os.Unsetenv(key)
}
