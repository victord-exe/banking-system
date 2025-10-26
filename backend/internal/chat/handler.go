package chat

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hlabs/banking-system/internal/middleware"
	"github.com/hlabs/banking-system/pkg/utils"
)

// Handler handles HTTP requests for chat operations
type Handler struct {
	service *Service
}

// NewHandler creates a new chat handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ProcessMessage handles chat message requests
// POST /api/chat
func (h *Handler) ProcessMessage(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse request body
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload. 'message' field is required")
		return
	}

	// Validate message is not empty
	if len(req.Message) == 0 {
		utils.RespondWithError(c, http.StatusBadRequest, "Message cannot be empty")
		return
	}

	// Validate message length (prevent abuse)
	if len(req.Message) > 500 {
		utils.RespondWithError(c, http.StatusBadRequest, "Message too long (maximum 500 characters)")
		return
	}

	// Process the message
	response, err := h.service.ProcessMessage(userID, req.Message)
	if err != nil {
		log.Printf("Error processing chat message for user %s: %v", userID, err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to process your request")
		return
	}

	// Return chat response
	utils.RespondWithSuccess(c, http.StatusOK, response, "Message processed successfully")
}
