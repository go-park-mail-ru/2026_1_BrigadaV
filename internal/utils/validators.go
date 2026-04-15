package utils

// import "regexp"

func IsValidEmail(email string) bool {
	// pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	// matched, _ := regexp.MatchString(pattern, email)
	// return matched
	return true
}

func IsValidNickname(nickname string) bool {
	return len(nickname) >= 3 && len(nickname) <= 50
}
