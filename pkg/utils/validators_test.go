package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidLogin(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		expected bool
	}{
		{"valid email", "user@example.com", true},
		{"valid with dots", "user.name@example.co.uk", true},
		{"valid with plus", "user+tag@example.com", true},
		{"valid with numbers", "user123@example.com", true},
		{"no @", "userexample.com", false},
		{"no domain", "user@", false},
		{"no tld", "user@example", false},
		{"empty", "", false},
		{"spaces", "user @example.com", false},
		{"special chars", "user!#$%@example.com", false},
		{"double @", "user@@example.com", false},
		{"unicode", "用户@example.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsValidLogin(tt.login))
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
			assert.Equal(t, tt.expected, IsValidNickname(tt.nickname))
		})
	}
}
