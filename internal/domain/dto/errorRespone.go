package dto

// swagger:model
type ErrorResponse struct {
	Errors string `json:"errors" example:"error description"`
}
