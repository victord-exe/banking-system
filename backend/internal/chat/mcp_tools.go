package chat

import (
	"context"
	"fmt"
	"strconv"
)

// handleGetBalance retrieves the current account balance for the authenticated user
// Returns balance in both cents (int64) and USD (float64)
//
// Expected args: none (userID comes from context)
// Returns: ToolResult with balance_cents and balance_usd in data
func (s *MCPServer) handleGetBalance(ctx context.Context, userID string, args map[string]interface{}) (ToolResult, error) {
	// Call account service to get balance in cents
	balanceCents, err := s.accountService.GetBalance(userID)
	if err != nil {
		return ToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to retrieve balance: %v", err),
		}, err
	}

	// Convert cents to USD
	balanceUSD := float64(balanceCents) / 100.0

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"balance_cents": balanceCents,
			"balance_usd":   balanceUSD,
		},
		Message: fmt.Sprintf("Current balance: $%.2f", balanceUSD),
	}, nil
}

// handleGetHistory retrieves transaction history for the authenticated user
// Supports optional "limit" parameter (default: 10, max: 100)
//
// Expected args:
//   - limit (optional): number of transactions to retrieve (default: 10, max: 100)
// Returns: ToolResult with transactions array and count
func (s *MCPServer) handleGetHistory(ctx context.Context, userID string, args map[string]interface{}) (ToolResult, error) {
	// Extract limit parameter (default: 10, max: 100)
	limit := 10
	if limitVal, ok := args["limit"]; ok {
		switch v := limitVal.(type) {
		case float64:
			limit = int(v)
		case int:
			limit = v
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				limit = parsed
			}
		}
	}

	// Validate limit bounds
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}

	// Call transaction service to get history (page 1)
	transactions, err := s.transactionService.GetHistory(userID, 1, limit)
	if err != nil {
		return ToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to retrieve transaction history: %v", err),
		}, err
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"transactions": transactions,
			"count":        len(transactions),
		},
		Message: fmt.Sprintf("Retrieved %d transactions", len(transactions)),
	}, nil
}

// handleDeposit adds funds to the user's account
// Amount is provided in USD and converted to cents internally
//
// Expected args:
//   - amount (float64): amount to deposit in USD (e.g., 100.50)
// Returns: ToolResult with success status
func (s *MCPServer) handleDeposit(ctx context.Context, userID string, args map[string]interface{}) (ToolResult, error) {
	// Extract amount from args (in USD)
	amountUSD, ok := args["amount"].(float64)
	if !ok {
		return ToolResult{
			Success: false,
			Message: "Invalid amount: must be a number",
		}, fmt.Errorf("invalid amount type")
	}

	// Validate amount is positive
	if amountUSD <= 0 {
		return ToolResult{
			Success: false,
			Message: "Invalid amount: must be greater than zero",
		}, fmt.Errorf("amount must be positive")
	}

	// Convert USD to cents
	amountCents := int64(amountUSD * 100)

	// Call transaction service to perform deposit
	err := s.transactionService.Deposit(userID, amountCents)
	if err != nil {
		return ToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to deposit funds: %v", err),
		}, err
	}

	return ToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully deposited $%.2f", amountUSD),
	}, nil
}

// handleWithdraw removes funds from the user's account
// Amount is provided in USD and converted to cents internally
// TigerBeetle automatically validates sufficient balance
//
// Expected args:
//   - amount (float64): amount to withdraw in USD (e.g., 50.00)
// Returns: ToolResult with success status
func (s *MCPServer) handleWithdraw(ctx context.Context, userID string, args map[string]interface{}) (ToolResult, error) {
	// Extract amount from args (in USD)
	amountUSD, ok := args["amount"].(float64)
	if !ok {
		return ToolResult{
			Success: false,
			Message: "Invalid amount: must be a number",
		}, fmt.Errorf("invalid amount type")
	}

	// Validate amount is positive
	if amountUSD <= 0 {
		return ToolResult{
			Success: false,
			Message: "Invalid amount: must be greater than zero",
		}, fmt.Errorf("amount must be positive")
	}

	// Convert USD to cents
	amountCents := int64(amountUSD * 100)

	// Call transaction service to perform withdrawal
	// Service will validate sufficient balance via TigerBeetle
	err := s.transactionService.Withdraw(userID, amountCents)
	if err != nil {
		return ToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to withdraw funds: %v", err),
		}, err
	}

	return ToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully withdrew $%.2f", amountUSD),
	}, nil
}

// handleTransfer sends funds from the user's account to another account
// Amount is provided in USD and converted to cents internally
// TigerBeetle validates sufficient balance and destination account existence
//
// Expected args:
//   - amount (float64): amount to transfer in USD (e.g., 75.50)
//   - to_account_id (string): destination TigerBeetle account ID (numeric string)
// Returns: ToolResult with success status
func (s *MCPServer) handleTransfer(ctx context.Context, userID string, args map[string]interface{}) (ToolResult, error) {
	// Extract amount from args (in USD)
	amountUSD, ok := args["amount"].(float64)
	if !ok {
		return ToolResult{
			Success: false,
			Message: "Invalid amount: must be a number",
		}, fmt.Errorf("invalid amount type")
	}

	// Validate amount is positive
	if amountUSD <= 0 {
		return ToolResult{
			Success: false,
			Message: "Invalid amount: must be greater than zero",
		}, fmt.Errorf("amount must be positive")
	}

	// Extract destination account ID
	toAccountIDStr, ok := args["to_account_id"].(string)
	if !ok {
		return ToolResult{
			Success: false,
			Message: "Invalid destination account: must be a string",
		}, fmt.Errorf("invalid to_account_id type")
	}

	// Parse destination account ID as uint64
	toAccountID, err := strconv.ParseUint(toAccountIDStr, 10, 64)
	if err != nil {
		return ToolResult{
			Success: false,
			Message: fmt.Sprintf("Invalid destination account ID: %v", err),
		}, err
	}

	// Convert USD to cents
	amountCents := int64(amountUSD * 100)

	// Call transaction service to perform transfer
	// Service will validate sufficient balance and destination account existence
	err = s.transactionService.Transfer(userID, toAccountID, amountCents)
	if err != nil {
		return ToolResult{
			Success: false,
			Message: fmt.Sprintf("Failed to transfer funds: %v", err),
		}, err
	}

	return ToolResult{
		Success: true,
		Message: fmt.Sprintf("Successfully transferred $%.2f to account %s", amountUSD, toAccountIDStr),
	}, nil
}

// initializeToolHandlers injects all handler functions into the MCPServer tools
// This function should be called during MCPServer initialization (in NewMCPServer)
//
// Parameters:
//   - server: The MCPServer instance to initialize
// Returns:
//   - error: If any handler registration fails
func initializeToolHandlers(server *MCPServer) error {
	// Register get_balance handler
	if err := server.SetToolHandler("get_balance", server.handleGetBalance); err != nil {
		return fmt.Errorf("failed to register get_balance handler: %w", err)
	}

	// Register get_transaction_history handler
	if err := server.SetToolHandler("get_transaction_history", server.handleGetHistory); err != nil {
		return fmt.Errorf("failed to register get_transaction_history handler: %w", err)
	}

	// Register deposit handler
	if err := server.SetToolHandler("deposit", server.handleDeposit); err != nil {
		return fmt.Errorf("failed to register deposit handler: %w", err)
	}

	// Register withdraw handler
	if err := server.SetToolHandler("withdraw", server.handleWithdraw); err != nil {
		return fmt.Errorf("failed to register withdraw handler: %w", err)
	}

	// Register transfer handler
	if err := server.SetToolHandler("transfer", server.handleTransfer); err != nil {
		return fmt.Errorf("failed to register transfer handler: %w", err)
	}

	return nil
}
