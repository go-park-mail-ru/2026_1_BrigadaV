package main

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSaltLengthAndUniqueness(t *testing.T) {
	salt1, err := generateSalt()
	require.NoError(t, err)
	salt2, err := generateSalt()
	require.NoError(t, err)

	assert.Len(t, salt1, saltLength)
	assert.Len(t, salt2, saltLength)
	assert.NotEqual(t, salt1, salt2, "salts should be unique")
}

func TestHashPasswordProducesValidFormat(t *testing.T) {
	pass := "mySecretPassword"
	hash, err := hashPassword(pass)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.True(t, strings.HasPrefix(hash, "argon2id$"), "hash should start with argon2id$")
	parts := strings.Split(hash, "$")
	assert.Len(t, parts, 5, "hash should have 5 parts separated by $")
}

func TestHashPasswordEmptyString(t *testing.T) {
	hash, err := hashPassword("")
	require.NoError(t, err)
	assert.NotEmpty(t, hash, "hash of empty string should not be empty")
}

func TestSamePasswordDifferentHashes(t *testing.T) {
	pass := "pass"
	hash1, err := hashPassword(pass)
	require.NoError(t, err)
	hash2, err := hashPassword(pass)
	require.NoError(t, err)
	assert.NotEqual(t, hash1, hash2, "hashes of same password should differ due to salt")
}

func TestCheckPasswordCorrect(t *testing.T) {
	pass := "pass123"
	hash, err := hashPassword(pass)
	require.NoError(t, err)
	ok, err := checkPassword(pass, hash)
	require.NoError(t, err)
	assert.True(t, ok, "checkPassword should return true for correct password")
}

func TestCheckPasswordWrong(t *testing.T) {
	pass := "pass123"
	hash, err := hashPassword(pass)
	require.NoError(t, err)
	ok, err := checkPassword("wrong", hash)
	require.NoError(t, err)
	assert.False(t, ok, "checkPassword should return false for wrong password")
}

func TestCheckPasswordInvalidHashFormat(t *testing.T) {
	_, err := checkPassword("pass", "invalid")
	assert.Error(t, err, "expected error for invalid hash format")
}

func TestCheckPasswordUnsupportedAlgorithm(t *testing.T) {
	_, err := checkPassword("pass", "md5$v=19$m=...")
	assert.Error(t, err, "expected error for unsupported algorithm")
}

func TestEncodeHashDecodable(t *testing.T) {
	salt := []byte("0123456789abcdef")
	hash := []byte("abcdefghijklmnopqrstuvwxyz012345")
	encoded := encodeHash(salt, hash)

	parts := strings.Split(encoded, "$")
	assert.Len(t, parts, 5, "encoded hash should have 5 parts")
	assert.Equal(t, "argon2id", parts[0], "algorithm should be argon2id")

	_, err := base64.RawStdEncoding.DecodeString(parts[3])
	assert.NoError(t, err, "salt part should be valid base64")
	_, err = base64.RawStdEncoding.DecodeString(parts[4])
	assert.NoError(t, err, "hash part should be valid base64")
}

func TestGenerateSessionTokenNotEmpty(t *testing.T) {
	token, err := generateSessionToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateSessionTokenLength(t *testing.T) {
	token, err := generateSessionToken()
	require.NoError(t, err)
	assert.Contains(t, []int{43, 44}, len(token), "token length should be 43 or 44")
}

func TestGenerateSessionTokenUniqueness(t *testing.T) {
	token1, err := generateSessionToken()
	require.NoError(t, err)
	token2, err := generateSessionToken()
	require.NoError(t, err)
	assert.NotEqual(t, token1, token2, "tokens should be unique")
}

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello@world", "@", true},
		{"helloworld", "@", false},
		{"", "@", false},
		{"hello", "", true},
		{"abc", "abc", true},
		{"abc", "d", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, contains(tt.s, tt.substr), "contains(%q, %q)", tt.s, tt.substr)
	}
}