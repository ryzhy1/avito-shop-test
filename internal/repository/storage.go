package repository

import (
	"errors"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrWrongPassword     = errors.New("wrong password")
)
