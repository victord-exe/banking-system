package account

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hlabs/banking-system/internal/middleware"
	"github.com/hlabs/banking-system/pkg/utils"
)

// Handler handles HTTP requests for account operations
type Handler struct {
	service *Service
}

// NewHandler creates a new account handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetAccountInfo returns the current user's account information
// GET /api/accounts/me
func (h *Handler) GetAccountInfo(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get user from database
	user, err := h.service.GetUserByID(userID)
	if err != nil {
		log.Printf("Error getting user info for %s: %v", userID, err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to retrieve account information")
		return
	}

	// Return user DTO (excludes password)
	utils.RespondWithSuccess(c, http.StatusOK, user.ToDTO(), "Account information retrieved successfully")
}

// GetBalance returns the current user's account balance
// GET /api/accounts/balance
func (h *Handler) GetBalance(c *gin.Context) {
	log.Printf("🔵 [AccountHandler] GetBalance called")

	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		log.Printf("❌ [AccountHandler] User not authenticated")
		utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	log.Printf("🔵 [AccountHandler] User ID from context: %s", userID)

	// Get balance from TigerBeetle
	log.Printf("🔵 [AccountHandler] Calling service.GetBalance for user %s...", userID)
	balance, err := h.service.GetBalance(userID)
	if err != nil {
		log.Printf("❌ [AccountHandler] Error getting balance for user %s: %v", userID, err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to retrieve balance")
		return
	}
	log.Printf("✅ [AccountHandler] Balance retrieved: %d cents for user %s", balance, userID)

	// Return balance as cents (TigerBeetle uses integer amounts)
	// Frontend should divide by 100 to get dollars
	response := gin.H{
		"balance":  balance,
		"currency": "USD",
	}

	log.Printf("✅ [AccountHandler] Sending response: %+v", response)
	utils.RespondWithSuccess(c, http.StatusOK, response, "Balance retrieved successfully")
}
