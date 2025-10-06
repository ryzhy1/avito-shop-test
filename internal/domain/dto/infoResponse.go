package dto

// swagger:model
type InfoResponse struct {
	Coins       int            `json:"coins" example:"100000"`
	Inventory   []PurchaseDTO  `json:"inventory"`
	CoinHistory TransactionDTO `json:"coin_history"`
}
