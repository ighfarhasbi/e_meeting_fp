package utils

import "fmt"

func ValidatePasswordCharacters(password string) error {
	if len(password) < 8 || !hasUppercase(password) || !hasLowercase(password) || !hasNumber(password) || !hasSpecialChar(password) {
		return fmt.Errorf("password must be at least 8 characters long and contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	}
	return nil
}

func hasUppercase(s string) bool {
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			return true
		}
	}
	return false
}

func hasLowercase(s string) bool {
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			return true
		}
	}
	return false
}

func hasNumber(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

func hasSpecialChar(s string) bool {
	specialChars := "!@#$%^&*()-_=+[]{}|;:',.<>?/"
	for _, c := range s {
		if contains(specialChars, c) {
			return true
		}
	}
	return false
}

func contains(s string, c rune) bool {
	for _, ch := range s {
		if ch == c {
			return true
		}
	}
	return false
}
