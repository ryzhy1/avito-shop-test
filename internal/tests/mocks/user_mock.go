package mocks

import (
	"avito-shop/internal/domain/dto"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) GetUserById(ctx context.Context, userID uuid.UUID) (dto.UserDTO, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(dto.UserDTO), args.Error(1)
}

func (m *UserRepositoryMock) GetUserPurchases(ctx context.Context, userID uuid.UUID) ([]dto.PurchaseDTO, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]dto.PurchaseDTO), args.Error(1)
}

func (m *UserRepositoryMock) GetCoinTransactions(ctx context.Context, userID uuid.UUID) (dto.TransactionDTO, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(dto.TransactionDTO), args.Error(1)
}

func (m *UserRepositoryMock) TransferCoins(ctx context.Context, fromUserID, toUserID uuid.UUID, amount int) error {
	args := m.Called(ctx, fromUserID, toUserID, amount)
	return args.Error(0)
}

func (m *UserRepositoryMock) BuyItem(ctx context.Context, userID uuid.UUID, item string) error {
	args := m.Called(ctx, userID, item)
	return args.Error(0)
}
