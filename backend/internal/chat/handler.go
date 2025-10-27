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

	// Validate message length (prevent abuse and API cost overruns)
	if len(req.Message) > 2000 {
		utils.RespondWithError(c, http.StatusBadRequest, "Message too long (maximum 2000 characters)")
		return
	}

	// Process the message
	response, err := h.service.ProcessMessage(userID, req.Message)
	if err != nil {
		log.Printf("❌ Error processing chat message for user %s: %v", userID, err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to process your request")
		return
	}

	// Log the complete response being sent to frontend
	log.Printf("📤 HANDLER: Sending response to frontend:")
	log.Printf("   Reply: %s", response.Reply)
	log.Printf("   Intent: %s", response.Intent)
	log.Printf("   RequiresConfirmation: %v", response.RequiresConfirmation)
	log.Printf("   Data: %+v", response.Data)
	if response.ConfirmationData != nil {
		log.Printf("   ConfirmationData.ToolName: %s", response.ConfirmationData.ToolName)
		log.Printf("   ConfirmationData.Arguments: %+v", response.ConfirmationData.Arguments)
	}

	// Return chat response
	utils.RespondWithSuccess(c, http.StatusOK, response, "Message processed successfully")
}

// ProcessConfirmation handles confirmation requests for critical operations
// POST /api/chat/confirm
func (h *Handler) ProcessConfirmation(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse request body
	var req ConfirmationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid confirmation request. 'tool_name', 'arguments', and 'confirmed' are required")
		return
	}

	// Validate tool name is not empty
	if len(req.ToolName) == 0 {
		utils.RespondWithError(c, http.StatusBadRequest, "Tool name cannot be empty")
		return
	}

	// Process the confirmation
	response, err := h.service.ProcessConfirmation(userID, req.ToolName, req.Arguments, req.Confirmed)
	if err != nil {
		log.Printf("Error processing confirmation for user %s: %v", userID, err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to process confirmation")
		return
	}

	// Return response
	utils.RespondWithSuccess(c, http.StatusOK, response, "Confirmation processed successfully")
}
