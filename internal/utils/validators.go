package utils

import "strings"

func IsValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
