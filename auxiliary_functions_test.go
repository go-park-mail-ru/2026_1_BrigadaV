package main

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestGenerateSaltLengthAndUniqueness(t *testing.T) {
	salt1, err1 := generateSalt()
	salt2, err2 := generateSalt()

	if err1 != nil || err2 != nil {
		t.Fatalf("generateSalt error: %v, %v", err1, err2)
	}
	if len(salt1) != saltLength {
		t.Errorf("expected salt length %d, got %d", saltLength, len(salt1))
	}
	if len(salt2) != saltLength {
		t.Errorf("expected salt length %d, got %d", saltLength, len(salt2))
	}
	if string(salt1) == string(salt2) {
		t.Error("two salts are identical, expected them to be different")
	}
}

func TestHashPasswordProducesValidFormat(t *testing.T) {
	pass := "mySecretPassword"
	hash, err := hashPassword(pass)
	if err != nil {
		t.Fatalf("hashPassword error: %v", err)
	}
	if hash == "" {
		t.Error("hashPassword returned empty string")
	}
	if !strings.HasPrefix(hash, "argon2id$") {
		t.Errorf("hash does not start with argon2id$: %s", hash)
	}
	parts := strings.Split(hash, "$")
	if len(parts) != 5 {
		t.Errorf("expected 5 parts, got %d", len(parts))
	}
	if len(hash) < 80 {
		t.Errorf("hash too short: %d", len(hash))
	}
}

func TestHashPasswordEmptyString(t *testing.T) {
	hash, err := hashPassword("")
	if err != nil {
		t.Fatalf("hashPassword with empty string error: %v", err)
	}
	if hash == "" {
		t.Error("hashPassword returned empty string for empty password")
	}
}

func TestSamePasswordDifferentHashes(t *testing.T) {
	pass := "pass"
	hash1, _ := hashPassword(pass)
	hash2, _ := hashPassword(pass)
	if hash1 == hash2 {
		t.Error("same password produced identical hash, expected different due to salt")
	}
}

func TestCheckPasswordCorrect(t *testing.T) {
	pass := "pass123"
	hash, _ := hashPassword(pass)
	ok, err := checkPassword(pass, hash)
	if err != nil {
		t.Fatalf("checkPassword error: %v", err)
	}
	if !ok {
		t.Error("checkPassword returned false for correct password")
	}
}

func TestCheckPasswordWrong(t *testing.T) {
	pass := "pass123"
	wrong := "wrong"
	hash, _ := hashPassword(pass)
	ok, err := checkPassword(wrong, hash)
	if err != nil {
		t.Fatalf("checkPassword error: %v", err)
	}
	if ok {
		t.Error("checkPassword returned true for wrong password")
	}
}

func TestCheckPasswordInvalidHashFormat(t *testing.T) {
	_, err := checkPassword("pass", "invalid")
	if err == nil {
		t.Error("expected error for invalid hash format, got nil")
	}
}

func TestCheckPasswordUnsupportedAlgorithm(t *testing.T) {
	_, err := checkPassword("pass", "md5$v=19$m=...")
	if err == nil {
		t.Error("expected error for unsupported algorithm, got nil")
	}
}

func TestEncodeHashDecodable(t *testing.T) {
	salt := []byte("0123456789abcdef")
	hash := []byte("abcdefghijklmnopqrstuvwxyz012345")
	encoded := encodeHash(salt, hash)

	parts := strings.Split(encoded, "$")
	if len(parts) != 5 {
		t.Fatalf("encoded hash has %d parts, expected 5", len(parts))
	}
	if parts[0] != "argon2id" {
		t.Errorf("algorithm = %s, want argon2id", parts[0])
	}
	_, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		t.Errorf("salt base64 decode error: %v", err)
	}
	_, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		t.Errorf("hash base64 decode error: %v", err)
	}
}

func TestGenerateSessionTokenNotEmpty(t *testing.T) {
	token, err := generateSessionToken()
	if err != nil {
		t.Fatalf("generateSessionToken error: %v", err)
	}
	if token == "" {
		t.Error("token is empty")
	}
}

func TestGenerateSessionTokenLength(t *testing.T) {
	token, _ := generateSessionToken()
	if len(token) < 40 || len(token) > 45 {
		t.Errorf("unexpected token length %d, expected around 43", len(token))
	}
}

func TestGenerateSessionTokenUniqueness(t *testing.T) {
	token1, _ := generateSessionToken()
	token2, _ := generateSessionToken()
	if token1 == token2 {
		t.Error("tokens are identical")
	}
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
		if got := contains(tt.s, tt.substr); got != tt.want {
			t.Errorf("contains(%q, %q) = %v; want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}
