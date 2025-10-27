package chat

import (
	"context"
	"fmt"
)

// ToolHandler is the function signature for MCP tool handlers
type ToolHandler func(
	ctx context.Context,
	userID string,
	args map[string]interface{},
) (ToolResult, error)

// ToolResult is returned by tool execution
type ToolResult struct {
	Success              bool                   `json:"success"`
	Data                 map[string]interface{} `json:"data,omitempty"`
	Message              string                 `json:"message"`
	RequiresConfirmation bool                   `json:"requires_confirmation,omitempty"`
	ConfirmationMessage  string                 `json:"confirmation_message,omitempty"`
	ToolName             string                 `json:"tool_name,omitempty"`
	Arguments            map[string]interface{} `json:"arguments,omitempty"`
}

// Tool represents an MCP tool with its metadata and handler
type Tool struct {
	Name                 string
	Description          string
	InputSchema          map[string]interface{}
	Handler              ToolHandler
	RequiresConfirmation bool
}

// GetConfirmationMessage generates a confirmation message for the tool
func (t *Tool) GetConfirmationMessage(args map[string]interface{}) string {
	// Default confirmation messages
	switch t.Name {
	case "deposit":
		if amount, ok := args["amount"].(float64); ok {
			return fmt.Sprintf("Do you want to deposit $%.2f?", amount)
		}
	case "withdraw":
		if amount, ok := args["amount"].(float64); ok {
			return fmt.Sprintf("Do you want to withdraw $%.2f?", amount)
		}
	case "transfer":
		if amount, ok := args["amount"].(float64); ok {
			if toAccount, ok := args["to_account_id"].(string); ok {
				return fmt.Sprintf("Do you want to transfer $%.2f to account %s?", amount, toAccount)
			}
		}
	}
	return "Please confirm this operation."
}

// OpenRouterRequest represents a request to OpenRouter API
type OpenRouterRequest struct {
	Model    string                   `json:"model"`
	Messages []OpenRouterMessage      `json:"messages"`
	Tools    []OpenRouterToolDef      `json:"tools,omitempty"`
	Stream   bool                     `json:"stream,omitempty"`
}

// OpenRouterMessage represents a message in the conversation
type OpenRouterMessage struct {
	Role       string                 `json:"role"`
	Content    string                 `json:"content,omitempty"`
	ToolCalls  []OpenRouterToolCall   `json:"tool_calls,omitempty"`
	ToolCallID string                 `json:"tool_call_id,omitempty"`
	Name       string                 `json:"name,omitempty"`
}

// OpenRouterToolDef represents a tool definition for OpenRouter
type OpenRouterToolDef struct {
	Type     string                 `json:"type"`
	Function OpenRouterFunctionDef  `json:"function"`
}

// OpenRouterFunctionDef represents a function definition
type OpenRouterFunctionDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// OpenRouterToolCall represents a tool call from the AI
type OpenRouterToolCall struct {
	ID       string                    `json:"id"`
	Type     string                    `json:"type"`
	Function OpenRouterFunctionCall    `json:"function"`
}

// OpenRouterFunctionCall represents the function call details
// Note: Arguments is a JSON string as per OpenRouter/OpenAI API spec
type OpenRouterFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string, not object
}

// OpenRouterResponse represents the response from OpenRouter
type OpenRouterResponse struct {
	ID      string                  `json:"id"`
	Object  string                  `json:"object"`
	Created int64                   `json:"created"`
	Model   string                  `json:"model"`
	Choices []OpenRouterChoice      `json:"choices"`
	Usage   OpenRouterUsage         `json:"usage"`
}

// OpenRouterChoice represents a choice in the response
type OpenRouterChoice struct {
	Index        int                `json:"index"`
	Message      OpenRouterMessage  `json:"message"`
	FinishReason string             `json:"finish_reason"`
}

// OpenRouterUsage represents token usage
type OpenRouterUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ProcessResult is the result of processing a user message through the AI client
// It includes the AI's reply and any confirmation metadata needed by the frontend
type ProcessResult struct {
	Reply                string                 // The natural language response from the AI
	RequiresConfirmation bool                   // Whether the operation requires user confirmation
	ToolName             string                 // Name of the tool that requires confirmation (if any)
	Arguments            map[string]interface{} // Arguments for the tool (if confirmation required)
	Data                 map[string]interface{} // Data returned by successful tool execution (balance, transactions, etc.)
}
