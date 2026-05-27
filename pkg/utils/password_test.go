package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("12345678")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Contains(t, hash, "argon2id$")
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestHashPassword_Different(t *testing.T) {
	hash1, _ := HashPassword("password1")
	hash2, _ := HashPassword("password2")
	assert.NotEqual(t, hash1, hash2)
}

func TestCheckPasswordHash_Correct(t *testing.T) {
	password := "12345678"
	hash, _ := HashPassword(password)
	assert.True(t, CheckPasswordHash(password, hash))
}

func TestCheckPasswordHash_Incorrect(t *testing.T) {
	password := "12345678"
	hash, _ := HashPassword(password)
	assert.False(t, CheckPasswordHash("wrongpassword", hash))
}

func TestCheckPasswordHash_InvalidHash(t *testing.T) {
	assert.False(t, CheckPasswordHash("password", "invalidhash"))
	assert.False(t, CheckPasswordHash("password", "argon2id$badformat"))
	assert.False(t, CheckPasswordHash("password", "argon2id$v=19$m=65536,t=1,p=4$salt$hash"))
}

func TestGenerateSalt(t *testing.T) {
	salt, err := generateSalt()
	assert.NoError(t, err)
	assert.Len(t, salt, saltLength)
}
