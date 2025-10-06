package handlers

import (
	"avito-shop/internal/domain/dto"
	"avito-shop/internal/services"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"time"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (accessToken string, refreshToken string, err error)
}

type AuthHandler struct {
	log         *slog.Logger
	authService AuthService
}

func NewAuthHandler(log *slog.Logger, authService AuthService) *AuthHandler {
	return &AuthHandler{
		log:         log,
		authService: authService,
	}
}

// Auth
// @Summary Аутентификация и получение JWT-токена
// @Description При первой аутентификации пользователь создается автоматически.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   auth body dto.AuthRequest true "Данные для аутентификации"
// @Success 200 {object} dto.AuthResponse "Успешная аутентификация"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 401 {object} dto.ErrorResponse "Неавторизован"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/auth [post]
func (h *AuthHandler) Auth(c *gin.Context) {
	var input dto.AuthRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.authService.Login(c.Request.Context(), input.Username, input.Password)
	if err != nil {
		if errors.Is(err, services.ErrFailedToGenerateTokens) || errors.Is(err, services.ErrFailedToStoreRefreshToken) {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		} else if errors.Is(err, services.ErrUserAlreadyExists) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid credentials"})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"message":      "Authorization successful",
		"token":        accessToken,
		"refreshToken": refreshToken,
		"time":         time.Now().Format(time.RFC3339),
	})
}
