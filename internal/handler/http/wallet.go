package http

import (
	"errors"
	"net/http"

	"github.com/AntonioForYou/wallet-api/internal/domain"
	"github.com/AntonioForYou/wallet-api/internal/repository/postgres"
	"github.com/AntonioForYou/wallet-api/internal/worker"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	pool *worker.Pool
	repo domain.WalletRepository
}

func NewHandler(pool *worker.Pool, repo domain.WalletRepository) *Handler {
	return &Handler{
		pool: pool,
		repo: repo,
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		v1.POST("/wallet", h.handleDepositWithdraw)
		v1.GET("/wallets/:walletId", h.handleGetBalance)
	}
}

func (h *Handler) handleDepositWithdraw(c *gin.Context) {
	var req domain.WalletRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body or missing fields"})
		return
	}

	resultChan := make(chan domain.Result, 1)

	job := domain.Job{
		Ctx:           c.Request.Context(),
		WalletID:      req.WalletID,
		OperationType: req.OperationType,
		Amount:        req.Amount,
		ResultChan:    resultChan,
	}

	h.pool.Dispatch(job)

	select {
	case <-c.Request.Context().Done():
		return
	case res := <-resultChan:
		if res.Err != nil {
			if errors.Is(res.Err, postgres.ErrWalletNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
				return
			}
			if errors.Is(res.Err, postgres.ErrInsufficientFunds) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient funds"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"walletId": req.WalletID,
			"balance":  res.NewBalance,
		})
	}
}

func (h *Handler) handleGetBalance(c *gin.Context) {
	walletIDStr := c.Param("walletId")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet id format"})
		return
	}

	balance, err := h.repo.GetBalance(c.Request.Context(), walletID)
	if err != nil {
		if errors.Is(err, postgres.ErrWalletNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"walletId": walletID,
		"balance":  balance,
	})
}
