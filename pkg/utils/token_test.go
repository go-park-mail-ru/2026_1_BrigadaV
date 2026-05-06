package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSessionToken(t *testing.T) {
	token, err := GenerateSessionToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Len(t, token, 44)
}

func TestGenerateSessionToken_Uniqueness(t *testing.T) {
	token1, _ := GenerateSessionToken()
	token2, _ := GenerateSessionToken()
	assert.NotEqual(t, token1, token2)
}

func TestHashToken(t *testing.T) {
	token := "test_token_123"
	hash := HashToken(token)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64)
}

func TestHashToken_Consistency(t *testing.T) {
	token := "test_token_123"
	hash1 := HashToken(token)
	hash2 := HashToken(token)
	assert.Equal(t, hash1, hash2)
}

func TestHashToken_DifferentInputs(t *testing.T) {
	hash1 := HashToken("token1")
	hash2 := HashToken("token2")
	assert.NotEqual(t, hash1, hash2)
}
