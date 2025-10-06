package unit

import (
	"avito-shop/internal/lib/jwt"
	"avito-shop/internal/middlewares"
	"avito-shop/internal/repository"
	"avito-shop/internal/services"
	"avito-shop/internal/tests/mocks"
	"context"
	"errors"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Login_RegistersNewUserAndStoresTokens(t *testing.T) {
	// Arrange
	ctx := context.Background()
	username := "newuser"
	password := "strongPass"

	authRepo := new(mocks.AuthRepositoryMock)
	redisMock := new(mocks.RedisClientMock)
	jwtGen := jwt.NewGenerator("secret", 0, 0)
	service := services.NewAuthService(slog.Default(), authRepo, redisMock, jwtGen)

	authRepo.On("LoginUser", ctx, "username", username).
		Return("", []byte{}, repository.ErrUserNotFound).Once()
	authRepo.On("SaveUser", ctx, username, mockHashedPassword(password)).
		Return(nil).Once()
	storedHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	authRepo.On("LoginUser", ctx, "username", username).
		Return("user-id", storedHash, nil).Once()
	redisMock.On("StoreRefreshToken", "user-id", mock.Anything).
		Return(nil).Once()

	// Act
	access, refresh, err := service.Login(ctx, username, password)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)
	authRepo.AssertExpectations(t)
	redisMock.AssertExpectations(t)
}

func TestAuthService_Login_ReturnsInvalidCredentialsForWrongPassword(t *testing.T) {
	// Arrange
	ctx := context.Background()
	username := "existing"
	password := "correctPass"

	authRepo := new(mocks.AuthRepositoryMock)
	redisMock := new(mocks.RedisClientMock)
	jwtGen := jwt.NewGenerator("secret", 0, 0)
	service := services.NewAuthService(slog.Default(), authRepo, redisMock, jwtGen)

	storedHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	authRepo.On("LoginUser", ctx, "username", username).
		Return("user-id", storedHash, nil).Once()

	// Act
	access, refresh, err := service.Login(ctx, username, "wrongPass")

	// Assert
	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
	assert.Empty(t, access)
	assert.Empty(t, refresh)
	authRepo.AssertExpectations(t)
	redisMock.AssertExpectations(t)
}

func TestAuthService_Login_PropagatesRepositoryErrors(t *testing.T) {
	// Arrange
	ctx := context.Background()
	username := "broken"

	authRepo := new(mocks.AuthRepositoryMock)
	redisMock := new(mocks.RedisClientMock)
	jwtGen := jwt.NewGenerator("secret", 0, 0)
	service := services.NewAuthService(slog.Default(), authRepo, redisMock, jwtGen)

	loginErr := errors.New("db failure")
	authRepo.On("LoginUser", ctx, "username", username).
		Return("", []byte{}, loginErr).Once()

	// Act
	access, refresh, err := service.Login(ctx, username, "password123")

	// Assert
	assert.ErrorContains(t, err, "db failure")
	assert.Empty(t, access)
	assert.Empty(t, refresh)
	authRepo.AssertExpectations(t)
	redisMock.AssertExpectations(t)
}

func TestAuthService_Login_ReturnsErrorWhenRefreshTokenStorageFails(t *testing.T) {
	// Arrange
	ctx := context.Background()
	username := "redisuser"
	password := "password123"

	authRepo := new(mocks.AuthRepositoryMock)
	redisMock := new(mocks.RedisClientMock)
	jwtGen := jwt.NewGenerator("secret", 0, 0)
	service := services.NewAuthService(slog.Default(), authRepo, redisMock, jwtGen)

	storedHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	authRepo.On("LoginUser", ctx, "username", username).
		Return("user-id", storedHash, nil).Once()
	redisErr := errors.New("redis down")
	redisMock.On("StoreRefreshToken", "user-id", mock.Anything).
		Return(redisErr).Once()

	// Act
	access, refresh, err := service.Login(ctx, username, password)

	// Assert
	assert.ErrorIs(t, err, services.ErrFailedToStoreRefreshToken)
	assert.Empty(t, access)
	assert.Empty(t, refresh)
	authRepo.AssertExpectations(t)
	redisMock.AssertExpectations(t)
}

func TestAuthService_Login_ReturnsErrorForInvalidInput(t *testing.T) {
	// Arrange
	authRepo := new(mocks.AuthRepositoryMock)
	redisMock := new(mocks.RedisClientMock)
	jwtGen := jwt.NewGenerator("secret", 0, 0)
	service := services.NewAuthService(slog.Default(), authRepo, redisMock, jwtGen)

	// Act
	access, refresh, err := service.Login(context.Background(), "", "short")

	// Assert
	assert.ErrorIs(t, err, middlewares.ErrEmptyField)
	assert.Empty(t, access)
	assert.Empty(t, refresh)
	authRepo.AssertNotCalled(t, "LoginUser", mock.Anything, mock.Anything, mock.Anything)
	redisMock.AssertNotCalled(t, "StoreRefreshToken", mock.Anything, mock.Anything)
}

func mockHashedPassword(password string) interface{} {
	return mock.MatchedBy(func(hash []byte) bool {
		return bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil
	})
}
