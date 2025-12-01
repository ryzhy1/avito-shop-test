package integration

import (
	"avito-shop/internal/domain/dto"
	"avito-shop/internal/lib/jwt"
	"avito-shop/internal/repository"
	"avito-shop/internal/services"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type memoryStorage struct {
	mu           sync.Mutex
	users        map[uuid.UUID]*userRecord
	purchases    []purchaseRecord
	transactions []transactionRecord
	merchPrices  map[string]int
}

type userRecord struct {
	username string
	password []byte
	coins    int
}

type purchaseRecord struct {
	userID uuid.UUID
	merch  string
}

type transactionRecord struct {
	from   uuid.UUID
	to     uuid.UUID
	amount int
}

func newMemoryStorage() *memoryStorage {
	return &memoryStorage{
		users:       make(map[uuid.UUID]*userRecord),
		merchPrices: defaultMerchPrices(),
	}
}

func defaultMerchPrices() map[string]int {
	return map[string]int{
		"t-shirt":    8000,
		"cup":        2000,
		"book":       5000,
		"pen":        1000,
		"powerbank":  20000,
		"hoody":      30000,
		"umbrella":   20000,
		"socks":      1000,
		"wallet":     5000,
		"pink-hoody": 50000,
	}
}

func (s *memoryStorage) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users = make(map[uuid.UUID]*userRecord)
	s.purchases = nil
	s.transactions = nil
}

func (s *memoryStorage) SaveUser(ctx context.Context, username string, passHash []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, user := range s.users {
		if user.username == username {
			return repository.ErrUserAlreadyExists
		}
	}

	id := uuid.New()
	s.users[id] = &userRecord{
		username: username,
		password: passHash,
		coins:    100000,
	}
	return nil
}

func (s *memoryStorage) LoginUser(ctx context.Context, inputType, input string) (string, []byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, user := range s.users {
		if inputType == "username" && user.username == input {
			return id.String(), user.password, nil
		}
	}

	return "", nil, repository.ErrUserNotFound
}

func (s *memoryStorage) CheckUsernameIsAvailable(ctx context.Context, username string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, user := range s.users {
		if user.username == username {
			return false, nil
		}
	}
	return true, nil
}

func (s *memoryStorage) GetUserById(ctx context.Context, userID uuid.UUID) (dto.UserDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return dto.UserDTO{}, repository.ErrUserNotFound
	}

	return dto.UserDTO{ID: userID, Username: user.username, Coins: user.coins}, nil
}

func (s *memoryStorage) GetUserPurchases(ctx context.Context, userID uuid.UUID) ([]dto.PurchaseDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	counts := make(map[string]int)
	for _, purchase := range s.purchases {
		if purchase.userID == userID {
			counts[purchase.merch]++
		}
	}

	var result []dto.PurchaseDTO
	for merch, amount := range counts {
		result = append(result, dto.PurchaseDTO{Merch: merch, Amount: amount})
	}

	return result, nil
}

func (s *memoryStorage) GetCoinTransactions(ctx context.Context, userID uuid.UUID) (dto.TransactionDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	received := make(map[uuid.UUID]int)
	sent := make(map[uuid.UUID]int)

	for _, tx := range s.transactions {
		if tx.to == userID {
			received[tx.from] += tx.amount
		}
		if tx.from == userID {
			sent[tx.to] += tx.amount
		}
	}

	toDTO := func(m map[uuid.UUID]int) []dto.CoinTransactionDTO {
		var res []dto.CoinTransactionDTO
		for id, total := range m {
			user := s.users[id]
			res = append(res, dto.CoinTransactionDTO{Username: user.username, TotalAmount: total})
		}
		return res
	}

	return dto.TransactionDTO{Received: toDTO(received), Sent: toDTO(sent)}, nil
}

func (s *memoryStorage) TransferCoins(ctx context.Context, fromUserID, toUserID uuid.UUID, amount int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fromUser, ok := s.users[fromUserID]
	if !ok {
		return repository.ErrUserNotFound
	}
	toUser, ok := s.users[toUserID]
	if !ok {
		return repository.ErrUserNotFound
	}

	if fromUser.coins < amount {
		return fmt.Errorf("insufficient funds")
	}

	fromUser.coins -= amount
	toUser.coins += amount
	s.transactions = append(s.transactions, transactionRecord{from: fromUserID, to: toUserID, amount: amount})
	return nil
}

func (s *memoryStorage) BuyItem(ctx context.Context, userID uuid.UUID, item string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return repository.ErrUserNotFound
	}
	price, ok := s.merchPrices[item]
	if !ok {
		return fmt.Errorf("item not found")
	}
	if user.coins < price {
		return fmt.Errorf("insufficient funds")
	}

	user.coins -= price
	s.purchases = append(s.purchases, purchaseRecord{userID: userID, merch: item})
	return nil
}

func (s *memoryStorage) Close() error { return nil }

type memoryRedis struct {
	mu    sync.Mutex
	store map[string]string
}

func newMemoryRedis() *memoryRedis {
	return &memoryRedis{store: make(map[string]string)}
}

func (r *memoryRedis) StoreRefreshToken(userID, refreshToken string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[refreshToken] = userID
	return nil
}

type IntegrationTestSuite struct {
	suite.Suite
	ctx          context.Context
	storage      *memoryStorage
	redisStorage *memoryRedis
	authService  *services.AuthService
	userService  *services.UserService
	jwtGen       *jwt.Generator
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.storage = newMemoryStorage()
	s.redisStorage = newMemoryRedis()
	s.jwtGen = jwt.NewGenerator("secret", time.Minute, 24*time.Hour)

	log := slog.Default()
	s.authService = services.NewAuthService(log, s.storage, s.redisStorage, s.jwtGen)
	s.userService = services.NewUserService(log, s.storage)
}

func (s *IntegrationTestSuite) SetupTest() {
	s.storage.reset()
	s.redisStorage = newMemoryRedis()
	log := slog.Default()
	s.authService = services.NewAuthService(log, s.storage, s.redisStorage, s.jwtGen)
	s.userService = services.NewUserService(log, s.storage)
}

func (s *IntegrationTestSuite) TestAuthLoginCreatesUserAndStoresRefreshToken() {
	access, refresh, err := s.authService.Login(s.ctx, "alice", "password123")
	s.Require().NoError(err)
	s.Require().NotEmpty(access)
	s.Require().NotEmpty(refresh)

	info, err := s.userService.GetUserInfo(s.ctx, s.findUserID("alice"))
	s.Require().NoError(err)
	s.Equal(100000, info.Coins)

	storedID := s.redisStorage.store[refresh]
	s.Equal(s.findUserID("alice").String(), storedID)
}

func (s *IntegrationTestSuite) TestAuthLoginWithWrongPassword() {
	s.createUser("bob", 100000, "secret123")

	access, refresh, err := s.authService.Login(s.ctx, "bob", "wrongpass")
	s.Require().Error(err)
	s.ErrorIs(err, services.ErrInvalidCredentials)
	s.Empty(access)
	s.Empty(refresh)
}

func (s *IntegrationTestSuite) TestUserInfoAggregatesPurchasesAndTransactions() {
	userID := s.createUser("user1", 100000, "pass")
	toUser := s.createUser("receiver", 100000, "pass")
	fromUser := s.createUser("sender", 100000, "pass")

	err := s.userService.TransferCoins(s.ctx, userID, toUser, 3000)
	s.Require().NoError(err)

	err = s.userService.TransferCoins(s.ctx, fromUser, userID, 5000)
	s.Require().NoError(err)

	err = s.userService.BuyItem(s.ctx, userID, "cup")
	s.Require().NoError(err)

	info, err := s.userService.GetUserInfo(s.ctx, userID)
	s.Require().NoError(err)

	s.Equal(100000, info.Coins)
	s.Len(info.Inventory, 1)
	s.Equal("cup", info.Inventory[0].Merch)
	s.Equal(1, info.Inventory[0].Amount)

	s.Len(info.CoinHistory.Sent, 1)
	s.Equal(3000, info.CoinHistory.Sent[0].TotalAmount)

	s.Len(info.CoinHistory.Received, 1)
	s.Equal(5000, info.CoinHistory.Received[0].TotalAmount)
}

func (s *IntegrationTestSuite) TestTransferCoinsUpdatesBalancesAndTransactions() {
	fromUser := s.createUser("from", 10000, "pass")
	toUser := s.createUser("to", 2000, "pass")

	err := s.userService.TransferCoins(s.ctx, fromUser, toUser, 3500)
	s.Require().NoError(err)

	fromInfo, err := s.userService.GetUserInfo(s.ctx, fromUser)
	s.Require().NoError(err)
	toInfo, err := s.userService.GetUserInfo(s.ctx, toUser)
	s.Require().NoError(err)

	s.Equal(6500, fromInfo.Coins)
	s.Equal(5500, toInfo.Coins)

	history, err := s.userService.GetCoinTransactions(s.ctx, fromUser)
	s.Require().NoError(err)
	s.Len(history.Sent, 1)
	s.Equal(3500, history.Sent[0].TotalAmount)
}

func (s *IntegrationTestSuite) TestBuyItemFailsWithInsufficientFunds() {
	userID := s.createUser("buyer", 500, "pass")

	err := s.userService.BuyItem(s.ctx, userID, "cup")
	s.Require().Error(err)

	info, err := s.userService.GetUserInfo(s.ctx, userID)
	s.Require().NoError(err)
	s.Equal(500, info.Coins)
	s.Empty(info.Inventory)
}

func (s *IntegrationTestSuite) createUser(username string, coins int, password string) uuid.UUID {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	s.Require().NoError(err)

	id := uuid.New()
	s.storage.mu.Lock()
	s.storage.users[id] = &userRecord{username: username, password: hash, coins: coins}
	s.storage.mu.Unlock()
	return id
}

func (s *IntegrationTestSuite) findUserID(username string) uuid.UUID {
	s.storage.mu.Lock()
	defer s.storage.mu.Unlock()
	for id, user := range s.storage.users {
		if user.username == username {
			return id
		}
	}
	return uuid.Nil
}
