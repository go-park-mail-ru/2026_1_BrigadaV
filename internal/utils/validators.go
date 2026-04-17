package utils

import "regexp"

func IsValidLogin(login string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, login)
	return matched
}

func IsValidNickname(nickname string) bool {
	return len(nickname) >= 3 && len(nickname) <= 50
}
