package handlers

import (
	"avito-shop/internal/domain/dto"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"time"
)

type UserService interface {
	GetUserPurchases(ctx context.Context, userID uuid.UUID) ([]dto.PurchaseDTO, error)
	GetCoinTransactions(ctx context.Context, userID uuid.UUID) (dto.TransactionDTO, error)
	TransferCoins(ctx context.Context, fromUserID, toUserID uuid.UUID, amount int) error
	GetUserInfo(ctx context.Context, userID uuid.UUID) (dto.InfoResponse, error)
	BuyItem(ctx context.Context, userID uuid.UUID, item string) error
}

type UserHandler struct {
	log         *slog.Logger
	userService UserService
}

func NewUserHandler(log *slog.Logger, userService UserService) *UserHandler {
	return &UserHandler{
		log:         log,
		userService: userService,
	}
}

// GetUserInfo godoc
// @Summary Получить информацию о монетах, инвентаре и истории транзакций
// @Description Возвращает баланс монет, инвентарь (купленные товары) и историю переводов монет.
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.InfoResponse "Информация о пользователе"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 401 {object} dto.ErrorResponse "Неавторизован"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/info [get]
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDVal.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	info, err := h.userService.GetUserInfo(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// TransferCoins
// @Summary Отправить монеты другому пользователю
// @Description Выполняет перевод монет от одного пользователя к другому.
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param transfer body dto.SendCoinRequest true "Данные для перевода"
// @Success 200 {string} string "Монеты успешно переведены"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 401 {object} dto.ErrorResponse "Неавторизован"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/sendCoins [post]
func (h *UserHandler) TransferCoins(c *gin.Context) {
	var input dto.SendCoinRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, _ := c.Get("user_id")

	if input.FromUserID.String() != userIDVal.(string) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You can't send coins from another user"})
		return
	}

	err := h.userService.TransferCoins(c.Request.Context(), input.FromUserID, input.ToUserID, input.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// BuyMerch
// @Summary Купить предмет за монеты
// @Description Покупает указанный предмет за монеты пользователя.
// @Tags user
// @Security BearerAuth
// @Produce json
// @Param item path string true "Название предмета"
// @Success 200 {string} string "Предмет куплен успешно"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 401 {object} dto.ErrorResponse "Неавторизован"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/buy/{item} [get]
func (h *UserHandler) BuyMerch(c *gin.Context) {
	item := c.Param("item")
	if item == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Item is required"})
		fmt.Println("hui")
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDVal.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		fmt.Println("zhopa")
		return
	}

	err = h.userService.BuyItem(c.Request.Context(), userID, item)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Item purchased successfully",
		"time":    time.Now().Format(time.RFC3339),
	})
}
