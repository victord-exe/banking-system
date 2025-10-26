package transaction

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hlabs/banking-system/internal/middleware"
	"github.com/hlabs/banking-system/pkg/utils"
)

// Handler handles HTTP requests for transaction operations
type Handler struct {
	service *Service
}

// NewHandler creates a new transaction handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// DepositRequest represents a deposit request payload
type DepositRequest struct {
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

// WithdrawRequest represents a withdrawal request payload
type WithdrawRequest struct {
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

// TransferRequest represents a transfer request payload
type TransferRequest struct {
	ToAccountID uint64 `json:"to_account_id" binding:"required"`
	Amount      int64  `json:"amount" binding:"required,gt=0"`
}

// Deposit handles deposit requests
// POST /api/transactions/deposit
func (h *Handler) Deposit(c *gin.Context) {
	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse request body
	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate amount (in cents, so max ~$10M)
	if req.Amount > 1000000000 {
		utils.RespondWithError(c, http.StatusBadRequest, "Deposit amount too large")
		return
	}

	// Execute deposit
	if err := h.service.Deposit(userID, req.Amount); err != nil {
		log.Printf("Deposit failed for user %s: %v", userID, err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to process deposit")
		return
	}

	response := gin.H{
		"amount":  req.Amount,
		"message": "Deposit successful",
	}

	utils.RespondWithSuccess(c, http.StatusOK, response, "Deposit completed successfully")
}

// Withdraw handles withdrawal requests
// POST /api/transactions/withdraw
func (h *Handler) Withdraw(c *gin.Context) {
	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse request body
	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Execute withdrawal
	if err := h.service.Withdraw(userID, req.Amount); err != nil {
		log.Printf("Withdrawal failed for user %s: %v", userID, err)

		// Check if it's an insufficient funds error
		if err.Error() == "insufficient funds" || len(err.Error()) > 18 && err.Error()[:18] == "insufficient funds" {
			utils.RespondWithError(c, http.StatusBadRequest, err.Error())
			return
		}

		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to process withdrawal")
		return
	}

	response := gin.H{
		"amount":  req.Amount,
		"message": "Withdrawal successful",
	}

	utils.RespondWithSuccess(c, http.StatusOK, response, "Withdrawal completed successfully")
}

// Transfer handles transfer requests
// POST /api/transactions/transfer
func (h *Handler) Transfer(c *gin.Context) {
	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse request body
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Execute transfer
	if err := h.service.Transfer(userID, req.ToAccountID, req.Amount); err != nil {
		log.Printf("Transfer failed from user %s to account %d: %v", userID, req.ToAccountID, err)

		// Check for specific errors
		errMsg := err.Error()
		if len(errMsg) > 18 && errMsg[:18] == "insufficient funds" {
			utils.RespondWithError(c, http.StatusBadRequest, errMsg)
			return
		}
		if errMsg == "destination account not found" {
			utils.RespondWithError(c, http.StatusNotFound, "Destination account does not exist")
			return
		}

		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to process transfer")
		return
	}

	response := gin.H{
		"to_account_id": req.ToAccountID,
		"amount":        req.Amount,
		"message":       "Transfer successful",
	}

	utils.RespondWithSuccess(c, http.StatusOK, response, "Transfer completed successfully")
}

// GetHistory handles transaction history requests
// GET /api/transactions/history?page=1&limit=10
func (h *Handler) GetHistory(c *gin.Context) {
	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse query parameters
	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if _, err := fmt.Sscanf(p, "%d", &page); err != nil || page < 1 {
			page = 1
		}
	}

	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil || limit < 1 || limit > 100 {
			limit = 10
		}
	}

	// Get transaction history
	history, err := h.service.GetHistory(userID, page, limit)
	if err != nil {
		log.Printf("Failed to get history for user %s: %v", userID, err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to retrieve transaction history")
		return
	}

	response := gin.H{
		"transactions": history,
		"page":         page,
		"limit":        limit,
	}

	utils.RespondWithSuccess(c, http.StatusOK, response, "Transaction history retrieved successfully")
}
