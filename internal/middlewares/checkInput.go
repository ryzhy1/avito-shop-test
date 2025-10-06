package middlewares

import (
	"fmt"
)

func CheckInput(login, password string) error {
	if login == "" || password == "" {
		return ErrEmptyField
	}

	if len(login) < 3 {
		return fmt.Errorf("%w: minimum 3 characters required", ErrLoginTooShort)
	}

	if len(password) < 8 {
		return fmt.Errorf("%w: minimum 8 characters required", ErrPasswordTooShort)
	}

	return nil
}
