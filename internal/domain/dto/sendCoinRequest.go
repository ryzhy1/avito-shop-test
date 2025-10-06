package dto

import "github.com/google/uuid"

// swagger:model
type SendCoinRequest struct {
	FromUserID uuid.UUID `json:"from_user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ToUserID   uuid.UUID `json:"to_user_id" example:"123e4567-e89b-12d3-a456-426614174001"`
	Amount     int       `json:"amount" example:"100"`
}
