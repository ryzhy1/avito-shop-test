package dto

type PurchaseDTO struct {
	Merch  string `json:"merch" db:"merch"`
	Amount int    `json:"amount" db:"amount"`
}
