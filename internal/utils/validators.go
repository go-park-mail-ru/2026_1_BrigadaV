package utils

import "strings"

func IsValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func IsValidPassword(password string) bool {
	return len(password) >= 8
}

func IsValidTitle(title string) bool {
	return len(title) >= 1 && len(title) <= 255
}
