package chat

// ChatRequest represents an incoming chat message from the user
type ChatRequest struct {
	Message string `json:"message" binding:"required"`
}

// ChatResponse represents the AI's response to a chat message
type ChatResponse struct {
	Reply                string                 `json:"reply"`
	Intent               Intent                 `json:"intent"`
	Data                 map[string]interface{} `json:"data,omitempty"`
	RequiresConfirmation bool                   `json:"requires_confirmation,omitempty"`
	ConfirmationData     *ConfirmationData      `json:"confirmation_data,omitempty"`
}

// ConfirmationData contains metadata needed for the frontend to confirm an operation
type ConfirmationData struct {
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Intent represents the detected user intent
type Intent string

const (
	IntentBalance  Intent = "balance"
	IntentDeposit  Intent = "deposit"
	IntentWithdraw Intent = "withdraw"
	IntentTransfer Intent = "transfer"
	IntentHistory  Intent = "history"
	IntentUnknown  Intent = "unknown"
)

// ParsedIntent contains the detected intent and extracted parameters
type ParsedIntent struct {
	Intent      Intent
	Amount      int64  // Amount in cents
	ToAccountID uint64 // For transfers
	Limit       int    // For history queries
}

// ConfirmationRequest represents a confirmation request for a critical operation
type ConfirmationRequest struct {
	ToolName  string                 `json:"tool_name" binding:"required"`
	Arguments map[string]interface{} `json:"arguments" binding:"required"`
	Confirmed bool                   `json:"confirmed" binding:"required"`
}
