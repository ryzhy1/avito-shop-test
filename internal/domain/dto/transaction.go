package dto

type TransactionDTO struct {
	Received []CoinTransactionDTO
	Sent     []CoinTransactionDTO
}
