package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "12345678"
	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Contains(t, hash, "argon2id$")
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("")

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestCheckPasswordHash_Correct(t *testing.T) {
	password := "12345678"
	hash, _ := HashPassword(password)

	result := CheckPasswordHash(password, hash)
	assert.True(t, result)
}

func TestCheckPasswordHash_Incorrect(t *testing.T) {
	password := "12345678"
	hash, _ := HashPassword(password)

	result := CheckPasswordHash("wrongpassword", hash)
	assert.False(t, result)
}

func TestCheckPasswordHash_InvalidHash(t *testing.T) {
	result := CheckPasswordHash("password", "invalidhash")
	assert.False(t, result)
}
