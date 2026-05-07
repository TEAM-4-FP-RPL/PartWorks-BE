package validator

import (
	"errors"
	"strings"
)

func ValidateEmail(email string) error {
	if strings.TrimSpace(email) == "" || !strings.Contains(email, "@") {
		return errors.New("invalid email")
	}
	return nil
}

func ValidatePassword(p string) error {
	if len(p) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}
