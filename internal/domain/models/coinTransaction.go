package models

import (
	"github.com/google/uuid"
	"time"
)

type CoinTransaction struct {
	ID         uuid.UUID `json:"id" db:"id"`
	FromUserID uuid.UUID `json:"from_user_id" db:"from_user_id"`
	ToUserID   uuid.UUID `json:"to_user_id" db:"to_user_id"`
	Amount     int       `json:"amount" db:"amount"` // хранится в копейках
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
