package dto

import (
	"github.com/google/uuid"
)

type UserDTO struct {
	ID       uuid.UUID `json:"id" db:"id"`
	Username string    `json:"username" db:"username"`
	Coins    int       `json:"coins" db:"coins"`
}
