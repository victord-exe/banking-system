package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// AIClient handles AI integration with OpenRouter for natural language banking operations
// It orchestrates the flow between user input, AI model, and MCP tool execution
type AIClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	mcpServer  *MCPServer
}

// NewAIClient creates a new AI client instance configured from environment variables
//
// Required environment variables:
//   - OPENROUTER_API_KEY: Your OpenRouter API key
//
// Optional environment variables:
//   - OPENROUTER_MODEL: AI model to use (default: "anthropic/claude-3.5-sonnet")
//   - OPENROUTER_BASE_URL: OpenRouter API base URL (default: "https://openrouter.ai/api/v1")
//
// Parameters:
//   - mcpServer: The MCP server instance that handles tool execution
//
// Returns:
//   - *AIClient: Configured AI client instance
//   - error: If API key is missing or invalid configuration
func NewAIClient(mcpServer *MCPServer) (*AIClient, error) {
	// Load API key (required)
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY environment variable is required")
	}

	// Load model (optional, with default)
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = "anthropic/claude-3.5-sonnet"
	}

	// Load base URL (optional, with default)
	baseURL := os.Getenv("OPENROUTER_BASE_URL")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}

	// Validate mcpServer is provided
	if mcpServer == nil {
		return nil, fmt.Errorf("mcpServer cannot be nil")
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &AIClient{
		apiKey:     apiKey,
		model:      model,
		baseURL:    baseURL,
		httpClient: httpClient,
		mcpServer:  mcpServer,
	}, nil
}

// ProcessMessage is the main orchestration method that handles natural language banking queries
//
// Flow:
//  1. Build OpenRouter request with system prompt, user message, and available tools
//  2. Call OpenRouter API to get AI response
//  3. Parse AI response:
//     - If text response → Return directly
//     - If tool call → Execute via MCP server
//  4. Handle confirmation flow:
//     - If tool requires confirmation → Return confirmation message with metadata
//     - If tool executed → Format and return result
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - userID: Authenticated user ID (from JWT)
//   - message: User's natural language message
//
// Returns:
//   - *ProcessResult: Result containing reply and confirmation metadata (if needed)
//   - error: Processing error if any
func (c *AIClient) ProcessMessage(ctx context.Context, userID, message string) (*ProcessResult, error) {
	// Validate inputs
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	// Trim and validate message
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, fmt.Errorf("message cannot be empty")
	}

	// Reasonable length limit to prevent abuse and excessive API costs
	if len(message) > 2000 {
		return nil, fmt.Errorf("message too long (maximum 2000 characters)")
	}

	// Step 1: Build OpenRouter request
	request := c.buildOpenRouterRequest(message)

	// Step 2: Call OpenRouter API
	response, err := c.callOpenRouter(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenRouter API: %w", err)
	}

	// Step 3: Handle AI response (pass original message for multi-turn conversation)
	return c.handleAIResponse(ctx, userID, message, response)
}

// buildOpenRouterRequest constructs an OpenRouter API request with banking context
//
// The request includes:
//   - System prompt: Establishes banking assistant persona and guidelines
//   - User message: The natural language query
//   - Tools: Available banking operations (from MCP server)
//
// Parameters:
//   - message: User's natural language message
//
// Returns:
//   - OpenRouterRequest: Configured API request
func (c *AIClient) buildOpenRouterRequest(message string) OpenRouterRequest {
	// Define system prompt with banking domain expertise
	systemPrompt := `You are a professional banking assistant. You help users with:
- Checking account balances
- Viewing transaction history
- Making deposits
- Making withdrawals
- Transferring money between accounts

Always be clear, professional, and security-conscious. For financial operations (deposit, withdraw, transfer), the system will automatically request confirmation before executing.

Use the provided tools to perform banking operations. Always provide friendly, natural language responses.

Important guidelines:
- For balance checks: Use get_balance tool
- For transaction history: Use get_transaction_history tool
- For deposits: Use deposit tool (requires confirmation)
- For withdrawals: Use withdraw tool (requires confirmation)
- For transfers: Use transfer tool (requires confirmation)

When users ask about operations in natural language, extract the relevant parameters:
- Amounts should be in USD (e.g., $100, 50 dollars, 25.50)
- Account IDs should be numeric strings
- Always confirm critical operations before execution`

	// Build messages array
	messages := []OpenRouterMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: message,
		},
	}

	// Get tool definitions from MCP server
	tools := c.mcpServer.GetToolDefinitions()

	// Log tool definitions being sent to OpenRouter
	log.Printf("🔧 TOOLS: Sending %d tools to OpenRouter:", len(tools))
	for i, tool := range tools {
		toolJSON, err := json.MarshalIndent(tool, "", "  ")
		if err != nil {
			log.Printf("   Tool %d: [Error marshaling: %v]", i+1, err)
			continue
		}
		log.Printf("   Tool %d:\n%s", i+1, string(toolJSON))
	}

	// Create OpenRouter request
	request := OpenRouterRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    tools,
		Stream:   false,
	}

	// Log the complete request being sent (without tool details since we logged them above)
	log.Printf("📤 OPENROUTER REQUEST:")
	log.Printf("   Model: %s", request.Model)
	log.Printf("   Messages: %d", len(request.Messages))
	log.Printf("   Tools: %d", len(request.Tools))
	log.Printf("   Stream: %v", request.Stream)

	return request
}

// callOpenRouter executes an HTTP request to the OpenRouter API
//
// Request flow:
//  1. Marshal request to JSON
//  2. Create HTTP POST request
//  3. Set required headers (Authorization, Content-Type, HTTP-Referer)
//  4. Execute request with context
//  5. Read and unmarshal response
//  6. Handle errors (non-200 status, API errors)
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - request: OpenRouter API request
//
// Returns:
//   - *OpenRouterResponse: API response with AI message and optional tool calls
//   - error: HTTP or API error if any
func (c *AIClient) callOpenRouter(ctx context.Context, request OpenRouterRequest) (*OpenRouterResponse, error) {
	// Marshal request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP POST request
	endpoint := fmt.Sprintf("%s/chat/completions", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set required headers
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("HTTP-Referer", "https://github.com/hlabs/banking-system")

	// Execute request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle non-200 status codes with sanitized error messages
	if httpResp.StatusCode != http.StatusOK {
		// Don't expose internal API errors to users - return sanitized messages
		switch httpResp.StatusCode {
		case http.StatusTooManyRequests:
			return nil, fmt.Errorf("AI service temporarily busy, please try again in a moment")
		case http.StatusUnauthorized, http.StatusForbidden:
			return nil, fmt.Errorf("AI service authentication error")
		case http.StatusBadRequest:
			return nil, fmt.Errorf("invalid request format")
		case http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return nil, fmt.Errorf("AI service temporarily unavailable")
		default:
			return nil, fmt.Errorf("AI service error (status %d)", httpResp.StatusCode)
		}
	}

	// Unmarshal response
	var response OpenRouterResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Log the response from OpenRouter
	log.Printf("📥 OPENROUTER RESPONSE:")
	log.Printf("   Choices: %d", len(response.Choices))
	if len(response.Choices) > 0 {
		choice := response.Choices[0]
		log.Printf("   Message Role: %s", choice.Message.Role)
		log.Printf("   Message Content: %s", choice.Message.Content)
		log.Printf("   Tool Calls: %d", len(choice.Message.ToolCalls))
		if len(choice.Message.ToolCalls) > 0 {
			log.Printf("   ✅ AI IS CALLING TOOLS!")
			for i, tc := range choice.Message.ToolCalls {
				log.Printf("      Tool Call %d: ID=%s, Type=%s, Function=%s", i+1, tc.ID, tc.Type, tc.Function.Name)
				log.Printf("      Arguments: %s", tc.Function.Arguments)
			}
		} else {
			log.Printf("   ❌ AI IS NOT CALLING ANY TOOLS - Just returning text")
		}
	}

	return &response, nil
}

// handleAIResponse processes the AI's response and executes tools if requested
//
// Response handling logic:
//  1. Check if response contains choices
//  2. Parse tool calls (if any)
//  3. If text response (no tool calls) → Return text directly
//  4. If tool call → Execute via MCP server
//  5. Handle confirmation flow:
//     - If requires confirmation → Return confirmation message with metadata
//     - If executed → Send tool result back to AI for natural language formatting
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - userID: Authenticated user ID
//   - userMessage: Original message from user (needed for multi-turn conversation)
//   - response: OpenRouter API response
//
// Returns:
//   - *ProcessResult: Result containing reply and confirmation metadata
//   - error: Processing error if any
func (c *AIClient) handleAIResponse(ctx context.Context, userID, userMessage string, response *OpenRouterResponse) (*ProcessResult, error) {
	// Validate response has choices
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in AI response")
	}

	choice := response.Choices[0]
	message := choice.Message

	// Check if AI wants to call a tool
	toolName, args, hasToolCall := c.parseToolCall(response)

	// Case 1: AI returns text response (no tool call)
	if !hasToolCall {
		reply := message.Content
		if reply == "" {
			reply = "I'm not sure how to help with that. Could you rephrase your question?"
		}
		return &ProcessResult{
			Reply:                reply,
			RequiresConfirmation: false,
		}, nil
	}

	// Case 2: AI wants to call a tool
	log.Printf("🔧 MCP_CLIENT: AI wants to call tool '%s' with args: %+v", toolName, args)

	// Execute tool via MCP server (with confirmed=false initially)
	result, err := c.mcpServer.ExecuteTool(ctx, toolName, userID, args, false)
	if err != nil {
		// Handle tool execution error
		log.Printf("❌ MCP_CLIENT: Tool execution error: %v", err)
		return &ProcessResult{
			Reply:                fmt.Sprintf("I encountered an error while processing your request: %v", err),
			RequiresConfirmation: false,
		}, nil
	}

	log.Printf("✅ MCP_CLIENT: Tool executed. Success=%v, RequiresConfirmation=%v", result.Success, result.RequiresConfirmation)
	log.Printf("   Message: %s", result.Message)
	log.Printf("   Data: %+v", result.Data)

	// Case 2a: Tool requires confirmation
	if result.RequiresConfirmation {
		log.Printf("⏸️  MCP_CLIENT: Tool requires confirmation, returning to user")
		return &ProcessResult{
			Reply:                result.ConfirmationMessage,
			RequiresConfirmation: true,
			ToolName:             toolName,
			Arguments:            args,
		}, nil
	}

	// Case 2b: Tool executed successfully - send result back to AI for natural language response
	log.Printf("🔄 MCP_CLIENT: Sending tool result back to AI for formatting")
	return c.handleToolResultWithAI(ctx, userMessage, response, toolName, result)
}

// handleToolResultWithAI sends tool execution results back to the AI for natural language formatting
//
// This implements a multi-turn conversation pattern:
//  1. User message → AI decides to call tool
//  2. Tool executes → Returns structured data
//  3. Tool result sent back to AI → AI formats as natural language
//
// This allows the AI to:
//   - Present data in a conversational, user-friendly way
//   - Add context and explanations
//   - Format numbers, dates, and complex data naturally
//   - Provide follow-up suggestions
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - userMessage: Original message from user
//   - originalResponse: The original AI response containing the tool call
//   - toolName: Name of the tool that was executed
//   - toolResult: Result from tool execution
//
// Returns:
//   - *ProcessResult: Final result with AI-formatted natural language response
//   - error: Processing error if any
func (c *AIClient) handleToolResultWithAI(ctx context.Context, userMessage string, originalResponse *OpenRouterResponse, toolName string, toolResult ToolResult) (*ProcessResult, error) {
	// Build conversation history with tool result
	// This follows OpenAI's function calling pattern
	
	// Get the original assistant message (with tool call)
	if len(originalResponse.Choices) == 0 {
		return nil, fmt.Errorf("no choices in original response")
	}
	
	assistantMessage := originalResponse.Choices[0].Message
	
	// Convert tool result to JSON string for the AI to read
	toolResultJSON, err := json.Marshal(toolResult.Data)
	if err != nil {
		// Fallback to simple message if data can't be marshaled
		toolResultJSON = []byte(fmt.Sprintf(`{"message": "%s", "success": %v}`, toolResult.Message, toolResult.Success))
	}
	
	// Build messages array with conversation history
	systemPrompt := `You are a professional banking assistant. Present the tool results in a natural, conversational way.

When presenting:
- Balance information: Show the amount clearly with currency formatting
- Transaction history: Summarize key details (date, type, amount, recipient)
- Be friendly and helpful
- Suggest relevant next actions when appropriate

Format monetary amounts as currency (e.g., $1,234.56).`

	messages := []OpenRouterMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userMessage, // Original user message
		},
		{
			Role:      "assistant",
			Content:   assistantMessage.Content,
			ToolCalls: assistantMessage.ToolCalls, // Include the tool call the AI made
		},
		{
			Role:       "tool",
			Content:    string(toolResultJSON),
			ToolCallID: assistantMessage.ToolCalls[0].ID, // Link to the tool call
			Name:       toolName,
		},
	}
	
	// Create new request with conversation history
	request := OpenRouterRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   false,
		// Don't include tools in this request - we want a text response, not another tool call
	}
	
	// Call OpenRouter again to get natural language response
	log.Printf("📞 MCP_CLIENT: Calling OpenRouter for AI formatting (2nd call)")
	response, err := c.callOpenRouter(ctx, request)
	if err != nil {
		// If AI formatting fails, fall back to basic formatted response
		log.Printf("⚠️  MCP_CLIENT: AI formatting failed, using fallback: %v", err)
		fallbackReply := c.formatToolResponse(toolResult)
		log.Printf("   Fallback reply: %s", fallbackReply)
		return &ProcessResult{
			Reply:                fallbackReply,
			RequiresConfirmation: false,
			Data:                 toolResult.Data,
		}, nil
	}

	// Extract the AI's natural language response
	if len(response.Choices) == 0 {
		log.Printf("⚠️  MCP_CLIENT: No choices in AI response, using fallback")
		fallbackReply := c.formatToolResponse(toolResult)
		return &ProcessResult{
			Reply:                fallbackReply,
			RequiresConfirmation: false,
			Data:                 toolResult.Data,
		}, nil
	}

	aiReply := response.Choices[0].Message.Content
	if aiReply == "" {
		// Fallback if AI doesn't provide content
		log.Printf("⚠️  MCP_CLIENT: AI returned empty content, using fallback")
		aiReply = c.formatToolResponse(toolResult)
	}

	log.Printf("✅ MCP_CLIENT: AI formatted response received")
	log.Printf("   AI Reply: %s", aiReply)
	log.Printf("   Data included: %+v", toolResult.Data)

	return &ProcessResult{
		Reply:                aiReply,
		RequiresConfirmation: false,
		Data:                 toolResult.Data, // Still include structured data for frontend
	}, nil
}

// parseToolCall extracts tool call information from OpenRouter response
//
// OpenRouter follows OpenAI's function calling format:
//   - Tool calls are in message.tool_calls array
//   - Each tool call has: id, type, function (name, arguments)
//   - Arguments is a JSON string that must be parsed
//
// Parameters:
//   - response: OpenRouter API response
//
// Returns:
//   - toolName: Name of the tool to execute
//   - args: Parsed tool arguments as map
//   - hasToolCall: True if tool call exists, false otherwise
func (c *AIClient) parseToolCall(response *OpenRouterResponse) (string, map[string]interface{}, bool) {
	log.Printf("🔍 PARSE_TOOL_CALL: Starting to parse tool call")

	// Validate response structure
	if len(response.Choices) == 0 {
		log.Printf("   ❌ No choices in response")
		return "", nil, false
	}

	message := response.Choices[0].Message
	log.Printf("   Message.ToolCalls length: %d", len(message.ToolCalls))

	// Check if tool_calls exist
	if len(message.ToolCalls) == 0 {
		log.Printf("   ❌ No tool calls in message")
		return "", nil, false
	}

	// Get first tool call
	toolCall := message.ToolCalls[0]
	toolName := toolCall.Function.Name
	argsString := toolCall.Function.Arguments

	log.Printf("   Tool Name: %s", toolName)
	log.Printf("   Arguments String: '%s' (length: %d)", argsString, len(argsString))

	var args map[string]interface{}

	// Handle empty arguments (valid for tools with no required parameters)
	// OpenRouter may return empty string "" or "{}" for parameterless tools
	trimmed := strings.TrimSpace(argsString)
	if trimmed == "" || trimmed == "{}" {
		log.Printf("   ⚠️  Arguments empty or '{}', using empty map (valid for parameterless tools)")
		return toolName, map[string]interface{}{}, true
	}

	// Parse non-empty arguments
	// OpenRouter returns arguments as JSON string: '{"amount": 100.50}'
	if err := json.Unmarshal([]byte(argsString), &args); err != nil {
		// Log error and return false to indicate parsing failure
		log.Printf("   ❌ JSON Unmarshal failed: %v", err)
		return toolName, nil, false
	}

	log.Printf("   ✅ Successfully parsed %d arguments", len(args))
	return toolName, args, true
}

// formatToolResponse converts ToolResult to natural language response
//
// Formatting logic:
//   - Success=false → Return error message
//   - Success=true → Return success message from tool handler
//
// The tool handlers in mcp_tools.go already provide user-friendly messages,
// so this method primarily handles error cases and ensures consistent formatting.
//
// Parameters:
//   - result: Tool execution result
//
// Returns:
//   - string: Natural language response
func (c *AIClient) formatToolResponse(result ToolResult) string {
	// Handle error case
	if !result.Success {
		return fmt.Sprintf("I encountered an error: %s", result.Message)
	}

	// Return success message
	// Tool handlers already provide natural language messages
	return result.Message
}

// GetModelInfo returns information about the configured AI model
// Useful for debugging and monitoring
//
// Returns:
//   - model: Model identifier (e.g., "anthropic/claude-3.5-sonnet")
//   - baseURL: API base URL
func (c *AIClient) GetModelInfo() (model, baseURL string) {
	return c.model, c.baseURL
}

// ValidateConfiguration checks if the client is properly configured
// Useful for health checks and initialization validation
//
// Returns:
//   - error: If configuration is invalid
func (c *AIClient) ValidateConfiguration() error {
	if c.apiKey == "" {
		return fmt.Errorf("API key is empty")
	}
	if c.model == "" {
		return fmt.Errorf("model is not configured")
	}
	if c.baseURL == "" {
		return fmt.Errorf("base URL is not configured")
	}
	if c.mcpServer == nil {
		return fmt.Errorf("MCP server is not initialized")
	}
	if c.httpClient == nil {
		return fmt.Errorf("HTTP client is not initialized")
	}
	return nil
}
