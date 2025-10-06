package services

import (
	"avito-shop/internal/lib/jwt"
	"avito-shop/internal/middlewares"
	"avito-shop/internal/repository"
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

type AuthService struct {
	log            *slog.Logger
	authRepository AuthRepository
	redis          RedisClient
	jwtGen         *jwt.Generator
}

type AuthRepository interface {
	SaveUser(ctx context.Context, login string, password []byte) error
	LoginUser(ctx context.Context, inputType, input string) (string, []byte, error)
	CheckUsernameIsAvailable(ctx context.Context, login string) (bool, error)
}

type RedisClient interface {
	StoreRefreshToken(userID, refreshToken string) error
}

var (
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrUserAlreadyExists         = errors.New("user already exists")
	ErrFailedToGenerateTokens    = errors.New("failed to generate tokens")
	ErrFailedToStoreRefreshToken = errors.New("failed to store refresh token")
)

func NewAuthService(log *slog.Logger, authRepository AuthRepository, redis RedisClient,
	jwtGen *jwt.Generator) *AuthService {
	return &AuthService{
		log:            log,
		authRepository: authRepository,
		redis:          redis,
		jwtGen:         jwtGen,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (accessToken string, refreshToken string,
	err error) {
	const op = "auth.Auth"

	log := s.log.With(
		slog.String("op", op),
		slog.String("username", username),
	)

	if err := middlewares.CheckInput(username, password); err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	id, storedHash, err := s.authRepository.LoginUser(ctx, "username", username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Info("user not found, registration")

			log.Info("hashing password")

			passHash, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if hashErr != nil {
				return "", "", fmt.Errorf("%s: %w", op, hashErr)
			}

			log.Info("password hashed")

			log.Info("saving user")

			err := s.authRepository.SaveUser(ctx, username, passHash)
			if err != nil {
				if errors.Is(err, repository.ErrUserAlreadyExists) {
					return "", "", fmt.Errorf("%s: %w", op, ErrUserAlreadyExists)
				}
				return "", "", fmt.Errorf("%s: %w", op, err)
			}

			log.Info("user saved")

			log.Info("login user")

			id, storedHash, err = s.authRepository.LoginUser(ctx, "username", username)
			if err != nil {
				return "", "", fmt.Errorf("%s: %w", op, err)
			}

			log.Info("user logged in")
		} else {
			log.Error("failed to login user", slog.String("error", err.Error()))
			return "", "", fmt.Errorf("%s: %w", op, err)
		}
	}

	log.Info("user found")

	log.Info("comparing passwords")

	err = bcrypt.CompareHashAndPassword(storedHash, []byte(password))
	if err != nil {
		if errors.Is(err, repository.ErrWrongPassword) {
			log.Info("invalid credentials", slog.String("error", err.Error()))
			return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("passwords match")

	log.Info("generating tokens")

	accessToken, refreshToken, err = s.jwtGen.GeneratePair(id)
	if err != nil {
		log.Error("failed to generate tokens", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, ErrFailedToGenerateTokens)
	}

	log.Info("tokens generated")

	log.Info("storing refresh token")

	if err := s.redis.StoreRefreshToken(id, refreshToken); err != nil {
		log.Error("failed to store refresh token", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, ErrFailedToStoreRefreshToken)
	}

	log.Info("refresh token stored")

	log.Info("tokens stored")

	return accessToken, refreshToken, nil
}
