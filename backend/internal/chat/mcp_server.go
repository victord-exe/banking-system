package chat

import (
	"context"
	"fmt"

	"github.com/hlabs/banking-system/internal/account"
	"github.com/hlabs/banking-system/internal/transaction"
)

// MCPServer is an embedded MCP server that exposes banking operations as tools
// It manages tool registration, execution, and confirmation flows for financial operations
type MCPServer struct {
	accountService     *account.Service
	transactionService *transaction.Service
	tools              map[string]*Tool
}

// NewMCPServer creates a new MCP server instance with banking tools
// Parameters:
//   - accountService: Service for account operations (balance queries)
//   - transactionService: Service for transaction operations (deposit, withdraw, transfer)
func NewMCPServer(accountService *account.Service, transactionService *transaction.Service) *MCPServer {
	server := &MCPServer{
		accountService:     accountService,
		transactionService: transactionService,
		tools:              make(map[string]*Tool),
	}

	// Register all banking tools
	server.registerTools()

	// Initialize tool handlers
	if err := initializeToolHandlers(server); err != nil {
		panic(fmt.Sprintf("Failed to initialize MCP tool handlers: %v", err))
	}

	return server
}

// registerTools registers all 5 banking tools with the MCP server
// Tools registered:
//   - get_balance: Query account balance (no confirmation)
//   - get_transaction_history: Query transaction history (no confirmation)
//   - deposit: Add funds to account (requires confirmation)
//   - withdraw: Remove funds from account (requires confirmation)
//   - transfer: Send funds to another account (requires confirmation)
func (s *MCPServer) registerTools() {
	// Tool 1: Get Balance
	s.tools["get_balance"] = &Tool{
		Name:        "get_balance",
		Description: "Get the current account balance for the authenticated user. Returns balance in USD and cents.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		},
		Handler:              nil, // Will be set in mcp_tools.go
		RequiresConfirmation: false,
	}

	// Tool 2: Get Transaction History
	s.tools["get_transaction_history"] = &Tool{
		Name:        "get_transaction_history",
		Description: "Retrieve recent transaction history for the authenticated user. Returns a paginated list of transactions.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Number of transactions to retrieve (default: 10, max: 100)",
					"default":     10,
					"minimum":     1,
					"maximum":     100,
				},
			},
			"required": []string{},
		},
		Handler:              nil, // Will be set in mcp_tools.go
		RequiresConfirmation: false,
	}

	// Tool 3: Deposit
	s.tools["deposit"] = &Tool{
		Name:        "deposit",
		Description: "Deposit money into the user's account. Requires confirmation before execution. Amount must be in USD (e.g., 100.50).",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"amount": map[string]interface{}{
					"type":        "number",
					"description": "Amount to deposit in USD (e.g., 100.50). Will be converted to cents internally.",
					"minimum":     0.01,
				},
			},
			"required": []string{"amount"},
		},
		Handler:              nil, // Will be set in mcp_tools.go
		RequiresConfirmation: true,
	}

	// Tool 4: Withdraw
	s.tools["withdraw"] = &Tool{
		Name:        "withdraw",
		Description: "Withdraw money from the user's account. Requires confirmation before execution. Amount must be in USD. Validates sufficient balance.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"amount": map[string]interface{}{
					"type":        "number",
					"description": "Amount to withdraw in USD (e.g., 50.00). Will be converted to cents internally.",
					"minimum":     0.01,
				},
			},
			"required": []string{"amount"},
		},
		Handler:              nil, // Will be set in mcp_tools.go
		RequiresConfirmation: true,
	}

	// Tool 5: Transfer
	s.tools["transfer"] = &Tool{
		Name:        "transfer",
		Description: "Transfer money to another account. Requires confirmation before execution. Validates destination account exists and sender has sufficient balance.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"amount": map[string]interface{}{
					"type":        "number",
					"description": "Amount to transfer in USD (e.g., 75.50). Will be converted to cents internally.",
					"minimum":     0.01,
				},
				"to_account_id": map[string]interface{}{
					"type":        "string",
					"description": "Destination TigerBeetle account ID (numeric string, e.g., '1761461878756072')",
				},
			},
			"required": []string{"amount", "to_account_id"},
		},
		Handler:              nil, // Will be set in mcp_tools.go
		RequiresConfirmation: true,
	}
}

// ExecuteTool executes a registered tool with confirmation flow handling
//
// Flow:
//   1. If tool requires confirmation AND confirmed=false → Return confirmation request
//   2. If tool requires confirmation AND confirmed=true → Execute tool handler
//   3. If tool doesn't require confirmation → Execute tool handler immediately
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - toolName: Name of the tool to execute
//   - userID: Authenticated user ID (from JWT)
//   - args: Tool arguments as key-value map
//   - confirmed: Whether user has confirmed the operation (for critical operations)
//
// Returns:
//   - ToolResult: Execution result or confirmation request
//   - error: Execution error if any
func (s *MCPServer) ExecuteTool(
	ctx context.Context,
	toolName string,
	userID string,
	args map[string]interface{},
	confirmed bool,
) (ToolResult, error) {
	// Validate tool exists
	tool, exists := s.tools[toolName]
	if !exists {
		return ToolResult{}, fmt.Errorf("tool '%s' not found", toolName)
	}

	// Validate handler is set
	if tool.Handler == nil {
		return ToolResult{}, fmt.Errorf("tool '%s' has no handler registered", toolName)
	}

	// Check if confirmation is required
	if tool.RequiresConfirmation && !confirmed {
		// Return confirmation request
		return ToolResult{
			Success:              false,
			RequiresConfirmation: true,
			ConfirmationMessage:  tool.GetConfirmationMessage(args),
			ToolName:             toolName,
			Arguments:            args,
			Message:              "Confirmation required for this operation",
		}, nil
	}

	// Execute the tool handler
	result, err := tool.Handler(ctx, userID, args)
	if err != nil {
		return ToolResult{
			Success: false,
			Message: fmt.Sprintf("Tool execution failed: %v", err),
		}, err
	}

	return result, nil
}

// GetToolDefinitions returns all registered tools in OpenRouter format
// This is used by the AI client to inform the LLM about available banking operations
//
// Returns:
//   - []OpenRouterToolDef: Array of tool definitions compatible with OpenRouter API
func (s *MCPServer) GetToolDefinitions() []OpenRouterToolDef {
	definitions := make([]OpenRouterToolDef, 0, len(s.tools))

	// Convert internal Tool format to OpenRouter format
	for _, tool := range s.tools {
		def := OpenRouterToolDef{
			Type: "function",
			Function: OpenRouterFunctionDef{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		}
		definitions = append(definitions, def)
	}

	return definitions
}

// GetTool retrieves a registered tool by name
// Used for testing and internal validation
//
// Parameters:
//   - name: Tool name to retrieve
//
// Returns:
//   - *Tool: The requested tool, or nil if not found
func (s *MCPServer) GetTool(name string) *Tool {
	return s.tools[name]
}

// SetToolHandler sets the handler function for a specific tool
// This allows handlers to be defined in separate files (mcp_tools.go)
//
// Parameters:
//   - toolName: Name of the tool to set handler for
//   - handler: Handler function to execute when tool is called
//
// Returns:
//   - error: If tool doesn't exist
func (s *MCPServer) SetToolHandler(toolName string, handler ToolHandler) error {
	tool, exists := s.tools[toolName]
	if !exists {
		return fmt.Errorf("tool '%s' not found", toolName)
	}

	tool.Handler = handler
	return nil
}
