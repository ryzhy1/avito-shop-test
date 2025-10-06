package tests

import (
	"avito-shop/internal/app"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type AuthResponse struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	Time         string `json:"time"`
}

func TestIntegration(t *testing.T) {
	t.Setenv("DB_NUMBER", "0")
	t.Setenv("REDIS_STORAGE_PATH", "localhost:6379")
	t.Setenv("redis_password", "somePass")

	logger := slog.Default()

	serverPort := "8080"
	storagePath := "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	secret := "secret_key"
	accessTTL := 15
	refreshTTL := 24

	application := app.New(
		logger,
		":"+serverPort,
		storagePath,
		secret,
		accessTTL,
		refreshTTL,
	)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := application.HTTPServer.Run(); err != nil {
			logger.Error("Server stopped with error", slog.String("error", err.Error()))
		}
	}()

	time.Sleep(1 * time.Second)

	baseURL := fmt.Sprintf("http://localhost:%s", serverPort)

	var authToken string

	t.Run("Ping_test", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/ping")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("Auth_test", func(t *testing.T) {
		body := `{"username": "testuser", "password": "Testpass123"}`
		resp, err := http.Post(baseURL+"/api/auth", "application/json", strToReadCloser(body))
		require.NoError(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200, got %d", resp.StatusCode)

		// Декодируем ответ
		var authResp AuthResponse
		err = json.NewDecoder(resp.Body).Decode(&authResp)
		require.NoError(t, err)
		require.NotEmpty(t, authResp.Token, "Token should not be empty")
		authToken = authResp.Token
	})

	t.Run("Info_test", func(t *testing.T) {
		req, err := http.NewRequest("GET", baseURL+"/api/info", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+authToken)

		cl := &http.Client{Timeout: 5 * time.Second}
		resp, err := cl.Do(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("BuyMerch_test", func(t *testing.T) {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/buy/%s", baseURL, "t-shirt"), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+authToken)
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("TransferCoins_test", func(t *testing.T) {
		fromUserID := "2084f5bf-bc61-4f0e-b3a4-7790b92114f7"
		toUserID := "d4cd5e68-43d2-468b-8303-7a25eb46c7fd"

		body := fmt.Sprintf(`{"from_user_id": "%s", "to_user_id": "%s", "amount": 50}`, fromUserID, toUserID)
		req, err := http.NewRequest("POST", baseURL+"/api/sendCoins", strToReadCloser(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken)
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})
}

func strToReadCloser(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}
