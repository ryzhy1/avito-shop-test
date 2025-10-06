package middlewares

import (
	"fmt"
)

func CheckRegister(login, email, password string) error {
	if login == "" || email == "" || password == "" {
		return ErrEmptyField
	}

	if !CorrectEmailChecker(email) {
		return ErrInvalidEmail
	}

	if len(login) < 3 {
		return fmt.Errorf("%w: minimum 3 characters required", ErrLoginTooShort)
	}

	if len(password) < 8 {
		return fmt.Errorf("%w: minimum 8 characters required", ErrPasswordTooShort)
	}

	return nil
}
