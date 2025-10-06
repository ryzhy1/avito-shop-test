package dto

type CoinTransactionDTO struct {
	Username    string `json:"username" db:"username"`
	TotalAmount int    `json:"total_amount" db:"total_amount"`
}
