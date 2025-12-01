package e2e

import (
	"avito-shop/internal/domain/dto"
	"avito-shop/internal/handlers"
	"avito-shop/internal/lib/jwt"
	"avito-shop/internal/middlewares"
	"avito-shop/internal/repository"
	"avito-shop/internal/routes"
	"avito-shop/internal/services"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
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
	for _, p := range s.purchases {
		if p.userID == userID {
			counts[p.merch]++
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

type testServer struct {
	server  *httptest.Server
	storage *memoryStorage
	jwtGen  *jwt.Generator
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()

	gin.SetMode(gin.TestMode)
	storage := newMemoryStorage()
	redisStorage := newMemoryRedis()
	jwtGen := jwt.NewGenerator("secret", time.Minute, 24*time.Hour)

	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	authService := services.NewAuthService(log, storage, redisStorage, jwtGen)
	userService := services.NewUserService(log, storage)

	authHandler := handlers.NewAuthHandler(log, authService)
	userHandler := handlers.NewUserHandler(log, userService)

	authMiddleware := middlewares.NewAuthMiddleware(jwtGen)
	router := routes.InitRoutes(authHandler, userHandler, authMiddleware)

	return &testServer{server: httptest.NewServer(router), storage: storage, jwtGen: jwtGen}
}

func (s *testServer) close() {
	s.server.Close()
}

func (s *testServer) url(path string) string {
	return s.server.URL + path
}

func (s *testServer) login(t *testing.T, username, password string) (token string, refresh string) {
	t.Helper()
	body := map[string]string{"username": username, "password": password}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	resp, err := http.Post(s.url("/api/auth"), "application/json", bytes.NewReader(payload))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refreshToken"`
	}
	err = json.NewDecoder(resp.Body).Decode(&parsed)
	require.NoError(t, err)

	return parsed.Token, parsed.RefreshToken
}

func (s *testServer) getInfo(t *testing.T, token string) dto.InfoResponse {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, s.url("/api/info"), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var info dto.InfoResponse
	err = json.NewDecoder(resp.Body).Decode(&info)
	require.NoError(t, err)
	return info
}

func (s *testServer) transferCoins(t *testing.T, token string, fromID, toID uuid.UUID, amount int) *http.Response {
	t.Helper()
	request := dto.SendCoinRequest{FromUserID: fromID, ToUserID: toID, Amount: amount}
	payload, err := json.Marshal(request)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, s.url("/api/sendCoins"), bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func (s *testServer) buy(t *testing.T, token string, item string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, s.url("/api/buy/"+item), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func (s *testServer) userIDByUsername(username string) uuid.UUID {
	s.storage.mu.Lock()
	defer s.storage.mu.Unlock()
	for id, user := range s.storage.users {
		if user.username == username {
			return id
		}
	}
	return uuid.Nil
}

func TestAuthAndInfoFlow(t *testing.T) {
	srv := newTestServer(t)
	defer srv.close()

	token, refresh := srv.login(t, "alice", "password123")
	require.NotEmpty(t, token)
	require.NotEmpty(t, refresh)

	info := srv.getInfo(t, token)
	require.Equal(t, 100000, info.Coins)
	require.Len(t, info.Inventory, 0)
	require.Len(t, info.CoinHistory.Received, 0)
	require.Len(t, info.CoinHistory.Sent, 0)
}

func TestTransferAndPurchaseFlow(t *testing.T) {
	srv := newTestServer(t)
	defer srv.close()

	aliceToken, _ := srv.login(t, "alice", "password123")
	bobToken, _ := srv.login(t, "bob", "password456")

	aliceID := srv.userIDByUsername("alice")
	bobID := srv.userIDByUsername("bob")

	resp := srv.transferCoins(t, aliceToken, aliceID, bobID, 5000)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp = srv.buy(t, bobToken, "cup")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	aliceInfo := srv.getInfo(t, aliceToken)
	require.Equal(t, 95000, aliceInfo.Coins)
	require.Len(t, aliceInfo.CoinHistory.Sent, 1)
	require.Equal(t, 5000, aliceInfo.CoinHistory.Sent[0].TotalAmount)

	bobInfo := srv.getInfo(t, bobToken)
	require.Equal(t, 103000, bobInfo.Coins)
	require.Len(t, bobInfo.Inventory, 1)
	require.Equal(t, "cup", bobInfo.Inventory[0].Merch)
	require.Len(t, bobInfo.CoinHistory.Received, 1)
	require.Equal(t, 5000, bobInfo.CoinHistory.Received[0].TotalAmount)
}
