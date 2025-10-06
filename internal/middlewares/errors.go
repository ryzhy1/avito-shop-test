package middlewares

import "errors"

var (
	ErrEmptyField       = errors.New("all fields must be filled")
	ErrInvalidEmail     = errors.New("email is invalid")
	ErrLoginTooShort    = errors.New("login must be at least 3 characters")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
)
