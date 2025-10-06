package models

import (
	"github.com/google/uuid"
	"time"
)

type Merch struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Merch     string    `json:"merch" db:"merch"`
	Price     int       `json:"price" db:"price"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
