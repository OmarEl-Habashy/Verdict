package validator

import (
	"errors"
	"strings"
)

// IsValidEmail validates email format (simplified)
func IsValidEmail(email string) (bool, error) {
	if email == "" {
		return false, errors.New("email is empty")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false, nil
	}

	if parts[0] == "" || parts[1] == "" {
		return false, nil
	}

	if !strings.Contains(parts[1], ".") {
		return false, nil
	}

	return true, nil
}

// IsValidPhoneNumber validates phone number format
func IsValidPhoneNumber(phone string) (bool, error) {
	if phone == "" {
		return false, errors.New("phone is empty")
	}

	// Remove common separators
	clean := strings.ReplaceAll(phone, "-", "")
	clean = strings.ReplaceAll(clean, " ", "")
	clean = strings.ReplaceAll(clean, "(", "")
	clean = strings.ReplaceAll(clean, ")", "")

	if len(clean) < 10 || len(clean) > 15 {
		return false, nil
	}

	// Check all characters are digits
	for _, ch := range clean {
		if ch < '0' || ch > '9' {
			return false, nil
		}
	}

	return true, nil
}

// IsValidPassword validates password strength
func IsValidPassword(password string) (bool, error) {
	if password == "" {
		return false, errors.New("password is empty")
	}

	if len(password) < 8 {
		return false, errors.New("password too short")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, ch := range password {
		if ch >= 'A' && ch <= 'Z' {
			hasUpper = true
		} else if ch >= 'a' && ch <= 'z' {
			hasLower = true
		} else if ch >= '0' && ch <= '9' {
			hasDigit = true
		} else if strings.ContainsRune("!@#$%^&*", ch) {
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return false, errors.New("password missing required characters")
	}

	return true, nil
}

// ValidateRange checks if value is within valid range
func ValidateRange(value int, min, max int) (bool, error) {
	if min > max {
		return false, errors.New("invalid range: min > max")
	}

	if value < min || value > max {
		return false, errors.New("value out of range")
	}

	return true, nil
}

// ValidateStringLength validates string is within length bounds
func ValidateStringLength(str string, minLen, maxLen int) (bool, error) {
	if minLen < 0 || maxLen < minLen {
		return false, errors.New("invalid length bounds")
	}

	length := len(str)
	if length < minLen || length > maxLen {
		return false, errors.New("string length out of bounds")
	}

	return true, nil
}

// IsValidUsername validates username format
func IsValidUsername(username string) (bool, error) {
	if username == "" {
		return false, errors.New("username is empty")
	}

	if len(username) < 3 || len(username) > 20 {
		return false, errors.New("username length invalid")
	}

	// Only alphanumeric and underscore allowed
	for _, ch := range username {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_') {
			return false, errors.New("invalid characters in username")
		}
	}

	return true, nil
}

// ValidateNotNil checks if value is not nil
func ValidateNotNil(value interface{}) (bool, error) {
	if value == nil {
		return false, errors.New("value is nil")
	}
	return true, nil
}
