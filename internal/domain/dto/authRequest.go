package dto

// swagger:model
type AuthRequest struct {
	Username string `json:"username" example:"johndoe"`
	Password string `json:"password" example:"secret"`
}
