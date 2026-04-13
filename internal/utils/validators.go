package utils

import "strings"

func IsValidNickname(nickname string) bool {
	return len(nickname) >= 3 && len(nickname) <= 50
}

func IsValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
