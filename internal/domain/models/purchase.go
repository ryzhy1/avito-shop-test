package models

import (
	"github.com/google/uuid"
	"time"
)

type Purchase struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	MerchID   uuid.UUID `json:"merch_id" db:"merch_id"`
	Amount    int       `json:"amount" db:"amount"` // храним в копейках если что чтбоы было проще хранить и перегонять в большую валюту
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
