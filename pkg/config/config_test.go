package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("JWT_SECRET", "a-very-long-secret-key-that-is-at-least-32-characters")
	os.Unsetenv("PORT")
	os.Unsetenv("FRONTEND_URL")
	os.Unsetenv("ALLOWED_ORIGINS")
	os.Unsetenv("SECURE_COOKIES")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "http://localhost:5173", cfg.FrontendURL)
	assert.False(t, cfg.SecureCookies)
}

func TestLoad_CustomPort(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("JWT_SECRET", "a-very-long-secret-key-that-is-at-least-32-characters")
	os.Setenv("PORT", "3000")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("PORT")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "3000", cfg.Port)
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Setenv("JWT_SECRET", "a-very-long-secret-key-that-is-at-least-32-characters")
	defer os.Unsetenv("JWT_SECRET")
	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL is required")
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("DATABASE_URL")
	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET is required")
}

func TestLoad_ShortJWTSecret(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("JWT_SECRET", "short")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("JWT_SECRET")
	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET must be at least 32 characters")
}

func TestLoad_AllowedOrigins(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("JWT_SECRET", "a-very-long-secret-key-that-is-at-least-32-characters")
	os.Setenv("ALLOWED_ORIGINS", "http://example.com, https://example.com")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.Len(t, cfg.AllowedOrigins, 2)
	assert.Equal(t, "http://example.com", cfg.AllowedOrigins[0])
	assert.Equal(t, "https://example.com", cfg.AllowedOrigins[1])
}

func TestLoad_AllowedOrigins_EmptyToken(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("JWT_SECRET", "a-very-long-secret-key-that-is-at-least-32-characters")
	os.Setenv("ALLOWED_ORIGINS", "http://example.com, , https://example.com")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.Len(t, cfg.AllowedOrigins, 2)
}

func TestLoad_SecureCookies(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("JWT_SECRET", "a-very-long-secret-key-that-is-at-least-32-characters")
	os.Setenv("SECURE_COOKIES", "true")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("SECURE_COOKIES")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.True(t, cfg.SecureCookies)
}

func TestGetEnv(t *testing.T) {
	os.Unsetenv("TEST_KEY")
	assert.Equal(t, "default", getEnv("TEST_KEY", "default"))
	os.Setenv("TEST_KEY", "custom")
	defer os.Unsetenv("TEST_KEY")
	assert.Equal(t, "custom", getEnv("TEST_KEY", "default"))
}

func TestGetEnvBool(t *testing.T) {
	os.Unsetenv("TEST_BOOL")
	assert.False(t, getEnvBool("TEST_BOOL", false))
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")
	assert.True(t, getEnvBool("TEST_BOOL", false))
	os.Setenv("TEST_BOOL", "1")
	assert.True(t, getEnvBool("TEST_BOOL", false))
	os.Setenv("TEST_BOOL", "yes")
	assert.True(t, getEnvBool("TEST_BOOL", false))
	os.Setenv("TEST_BOOL", "no")
	assert.False(t, getEnvBool("TEST_BOOL", false))
}
