package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents an error response structure
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a success response structure
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// RespondWithError sends an error JSON response
func RespondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
	})
}

// RespondWithSuccess sends a success JSON response
func RespondWithSuccess(c *gin.Context, code int, data interface{}, message ...string) {
	response := SuccessResponse{
		Data: data,
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	c.JSON(code, response)
}

// RespondWithData sends a JSON response with just the data
func RespondWithData(c *gin.Context, code int, data interface{}) {
	c.JSON(code, data)
}
