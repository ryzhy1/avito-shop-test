package services

import (
	"avito-shop/internal/domain/dto"
	"context"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
)

type UserService struct {
	log            *slog.Logger
	userRepository UserRepository
}

type UserRepository interface {
	GetUserById(ctx context.Context, userID uuid.UUID) (dto.UserDTO, error)
	GetUserPurchases(ctx context.Context, userID uuid.UUID) ([]dto.PurchaseDTO, error)
	GetCoinTransactions(ctx context.Context, userID uuid.UUID) (dto.TransactionDTO, error)
	TransferCoins(ctx context.Context, fromUserID, toUserID uuid.UUID, amount int) error
	BuyItem(ctx context.Context, userID uuid.UUID, item string) error
}

func NewUserService(log *slog.Logger, userRepository UserRepository) *UserService {
	return &UserService{
		log:            log,
		userRepository: userRepository,
	}
}

func (s *UserService) GetUserInfo(ctx context.Context, userID uuid.UUID) (dto.InfoResponse, error) {
	const op = "services.UserService.GetUserInfo"

	user, err := s.userRepository.GetUserById(ctx, userID)
	if err != nil {
		return dto.InfoResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	inventory, err := s.userRepository.GetUserPurchases(ctx, userID)
	if err != nil {
		return dto.InfoResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	coinHistory, err := s.userRepository.GetCoinTransactions(ctx, userID)
	if err != nil {
		return dto.InfoResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	return dto.InfoResponse{
		Coins:       user.Coins,
		Inventory:   inventory,
		CoinHistory: coinHistory,
	}, nil
}

func (s *UserService) GetUserPurchases(ctx context.Context, userID uuid.UUID) ([]dto.PurchaseDTO, error) {
	const op = "services.UserService.GetUserPurchases"

	log := s.log.With(
		slog.String("op", op),
		slog.String("user_id", userID.String()),
	)

	log.Info("getting user purchases")

	purchases, err := s.userRepository.GetUserPurchases(ctx, userID)
	if err != nil {
		log.Error("failed to get user purchases", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("got user purchases")

	return purchases, err
}

func (s *UserService) GetCoinTransactions(ctx context.Context, userID uuid.UUID) (dto.TransactionDTO, error) {
	const op = "services.UserService.GetCoinTransactions"

	log := s.log.With(
		slog.String("op", op),
		slog.String("user_id", userID.String()),
	)

	log.Info("getting user coin transactions")

	coinTransactions, err := s.userRepository.GetCoinTransactions(ctx, userID)
	if err != nil {
		log.Error("failed to get user coin transactions", err)
		return dto.TransactionDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("got user coin transactions")

	return coinTransactions, err
}

func (s *UserService) TransferCoins(ctx context.Context, fromUserID, toUserID uuid.UUID, amount int) error {
	const op = "services.UserService.TransferCoins"

	log := s.log.With(
		slog.String("op", op),
		slog.String("from_user_id", fromUserID.String()),
		slog.String("to_user_id", toUserID.String()),
		slog.Int("amount", amount),
	)

	log.Info("sending coins")

	if err := s.userRepository.TransferCoins(ctx, fromUserID, toUserID, amount); err != nil {
		log.Error("failed to transfer coins", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("coins sent")

	return nil
}

func (s *UserService) BuyItem(ctx context.Context, userID uuid.UUID, item string) error {
	const op = "services.UserService.BuyItem"

	log := s.log.With(
		slog.String("op", op),
		slog.String("user_id", userID.String()),
		slog.String("item", item),
	)

	log.Info("buying item")

	if err := s.userRepository.BuyItem(ctx, userID, item); err != nil {
		log.Error("failed to buy item", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("bought item")

	return nil
}
