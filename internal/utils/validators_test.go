package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"valid email", "user@example.com", true},
		{"valid with dots", "user.name@example.co.uk", true},
		{"no @", "userexample.com", false},
		{"no domain", "user@", false},
		{"empty", "", false},
		{"spaces", "user @example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidNickname(t *testing.T) {
	tests := []struct {
		name     string
		nickname string
		expected bool
	}{
		{"valid", "johnny", true},
		{"too short", "ab", false},
		{"too long", strings.Repeat("a", 51), false},
		{"empty", "", false},
		{"exactly 3 chars", "abc", true},
		{"exactly 50 chars", strings.Repeat("a", 50), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidNickname(tt.nickname)
			assert.Equal(t, tt.expected, result)
		})
	}
}
