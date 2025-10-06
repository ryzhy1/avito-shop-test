package unit

import (
	"avito-shop/internal/domain/dto"
	"avito-shop/internal/services"
	"avito-shop/internal/tests/mocks"
	"context"
	"errors"
	"testing"

	"log/slog"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserService_GetUserInfo_ReturnsAggregatedData(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := uuid.New()

	repo := new(mocks.UserRepositoryMock)
	repo.On("GetUserById", ctx, userID).
		Return(dto.UserDTO{Coins: 5000}, nil).Once()
	repo.On("GetUserPurchases", ctx, userID).
		Return([]dto.PurchaseDTO{{Merch: "pen", Amount: 2}}, nil).Once()
	repo.On("GetCoinTransactions", ctx, userID).
		Return(dto.TransactionDTO{Received: []dto.CoinTransactionDTO{{Username: "alice", TotalAmount: 100}}}, nil).Once()

	service := services.NewUserService(slog.Default(), repo)

	// Act
	info, err := service.GetUserInfo(ctx, userID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 5000, info.Coins)
	assert.Len(t, info.Inventory, 1)
	assert.Equal(t, "pen", info.Inventory[0].Merch)
	assert.NotEmpty(t, info.CoinHistory.Received)
	repo.AssertExpectations(t)
}

func TestUserService_GetUserInfo_ReturnsErrorWhenUserLookupFails(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := uuid.New()
	repoErr := errors.New("db down")

	repo := new(mocks.UserRepositoryMock)
	repo.On("GetUserById", ctx, userID).
		Return(dto.UserDTO{}, repoErr).Once()

	service := services.NewUserService(slog.Default(), repo)

	// Act
	info, err := service.GetUserInfo(ctx, userID)

	// Assert
	assert.ErrorContains(t, err, "db down")
	assert.Empty(t, info.Inventory)
	repo.AssertExpectations(t)
}

func TestUserService_GetUserInfo_StopsWhenPurchasesLookupFails(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := uuid.New()
	repo := new(mocks.UserRepositoryMock)

	repo.On("GetUserById", ctx, userID).
		Return(dto.UserDTO{Coins: 100}, nil).Once()
	repo.On("GetUserPurchases", ctx, userID).
		Return([]dto.PurchaseDTO(nil), errors.New("purchases error")).Once()

	service := services.NewUserService(slog.Default(), repo)

	// Act
	info, err := service.GetUserInfo(ctx, userID)

	// Assert
	assert.ErrorContains(t, err, "purchases error")
	assert.Zero(t, info.Coins)
	repo.AssertExpectations(t)
}

func TestUserService_TransferCoins_ReturnsErrorOnFailure(t *testing.T) {
	// Arrange
	ctx := context.Background()
	fromID := uuid.New()
	toID := uuid.New()
	repoErr := errors.New("transfer failed")

	repo := new(mocks.UserRepositoryMock)
	repo.On("TransferCoins", ctx, fromID, toID, 100).
		Return(repoErr).Once()

	service := services.NewUserService(slog.Default(), repo)

	// Act
	err := service.TransferCoins(ctx, fromID, toID, 100)

	// Assert
	assert.ErrorContains(t, err, "transfer failed")
	repo.AssertExpectations(t)
}

func TestUserService_BuyItem_PropagatesRepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := uuid.New()
	repoErr := errors.New("out of stock")

	repo := new(mocks.UserRepositoryMock)
	repo.On("BuyItem", ctx, userID, "t-shirt").
		Return(repoErr).Once()

	service := services.NewUserService(slog.Default(), repo)

	// Act
	err := service.BuyItem(ctx, userID, "t-shirt")

	// Assert
	assert.ErrorContains(t, err, "out of stock")
	repo.AssertExpectations(t)
}
