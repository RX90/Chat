package handler

import (
	"errors"
	"regexp"
	"unicode/utf8"

	"github.com/RX90/Chat/internal/domain/dto"
)

func ValidateEmail(email string) bool {
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(regex)
	return re.MatchString(email)
}

func ValidateUsername(username string) bool {
	return utf8.RuneCountInString(username) >= 4 && utf8.RuneCountInString(username) <= 32
}

func ValidatePassword(password string) bool {
	return utf8.RuneCountInString(password) >= 8 && utf8.RuneCountInString(password) <= 32
}

func inputValidation(input dto.SignUpUser) error {
	if !ValidateEmail(input.Email) {
		return errors.New("invalid email")
	}

	if !ValidateUsername(input.Username) {
		return errors.New("username must be 4-32 characters")
	}

	if !ValidatePassword(input.Password) {
		return errors.New("password must be 8-64 characters, with uppercase, lowercase, number and special character")
	}
	return nil
}
