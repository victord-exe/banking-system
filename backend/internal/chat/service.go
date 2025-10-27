package chat

import (
	"context"
	"fmt"
	"log"

	"github.com/hlabs/banking-system/internal/account"
	"github.com/hlabs/banking-system/internal/transaction"
)

// Service handles chat-related business logic
type Service struct {
	accountService     *account.Service
	transactionService *transaction.Service
	mcpServer          *MCPServer
	aiClient           *AIClient
}

// NewService creates a new chat service with MCP integration
func NewService(accountService *account.Service, transactionService *transaction.Service) *Service {
	// Initialize MCP Server with banking tools
	mcpServer := NewMCPServer(accountService, transactionService)
	log.Println("✅ MCP Server initialized with 5 banking tools")

	// Initialize AI Client (loads from environment variables)
	aiClient, err := NewAIClient(mcpServer)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to initialize AI Client: %v", err)
		log.Println("   Chat feature will be degraded. Check OPENROUTER_API_KEY environment variable.")
		// Continue with nil aiClient - service will handle gracefully
	} else {
		model, baseURL := aiClient.GetModelInfo()
		log.Printf("✅ AI Client initialized (Model: %s, URL: %s)", model, baseURL)
	}

	return &Service{
		accountService:     accountService,
		transactionService: transactionService,
		mcpServer:          mcpServer,
		aiClient:           aiClient,
	}
}

// ProcessMessage processes a user's chat message using AI and MCP tools
func (s *Service) ProcessMessage(userID, message string) (ChatResponse, error) {
	// Log chat interaction for audit
	log.Printf("💬 Chat: User %s: %s", userID, message)

	// Check if AI client is available
	if s.aiClient == nil {
		log.Printf("⚠️  AI Client not available, falling back to error response")
		return ChatResponse{
			Reply:                "AI chat service is currently unavailable. Please contact support.",
			Intent:               IntentUnknown,
			RequiresConfirmation: false,
		}, nil
	}

	// Process message through AI client
	ctx := context.Background()
	result, err := s.aiClient.ProcessMessage(ctx, userID, message)
	if err != nil {
		log.Printf("❌ Error processing message with AI: %v", err)
		return ChatResponse{}, fmt.Errorf("failed to process message: %w", err)
	}

	// Log what we received from AI client
	log.Printf("📥 SERVICE: Received ProcessResult from AI Client:")
	log.Printf("   Reply: %s", result.Reply)
	log.Printf("   RequiresConfirmation: %v", result.RequiresConfirmation)
	log.Printf("   ToolName: %s", result.ToolName)
	log.Printf("   Data: %+v", result.Data)

	// Build response with the AI's reply and confirmation metadata
	response := ChatResponse{
		Reply:                result.Reply,
		Intent:               IntentUnknown, // AI handles intent internally
		RequiresConfirmation: result.RequiresConfirmation,
		Data:                 result.Data, // Pass structured data from tools (balance, transactions, etc.)
	}

	// If confirmation is required, include the confirmation data
	if result.RequiresConfirmation {
		response.ConfirmationData = &ConfirmationData{
			ToolName:  result.ToolName,
			Arguments: result.Arguments,
		}
		log.Printf("💬 Chat: Confirmation required for %s", result.ToolName)
	} else {
		log.Printf("💬 Chat: AI Response delivered successfully")
	}

	return response, nil
}

// ProcessConfirmation processes a user's confirmation for a critical operation
func (s *Service) ProcessConfirmation(userID, toolName string, args map[string]interface{}, confirmed bool) (ChatResponse, error) {
	// Log confirmation interaction for audit
	log.Printf("💬 Confirmation: User %s - Tool: %s, Confirmed: %v", userID, toolName, confirmed)

	// Check if MCP server is available
	if s.mcpServer == nil {
		log.Printf("❌ MCP Server not available")
		return ChatResponse{}, fmt.Errorf("MCP server not initialized")
	}

	// If user declined, return appropriate message
	if !confirmed {
		return ChatResponse{
			Reply:                "Operation cancelled.",
			Intent:               IntentUnknown,
			RequiresConfirmation: false,
		}, nil
	}

	// Execute the tool with confirmation
	ctx := context.Background()
	result, err := s.mcpServer.ExecuteTool(ctx, toolName, userID, args, true)
	if err != nil {
		log.Printf("❌ Error executing confirmed tool %s: %v", toolName, err)
		return ChatResponse{}, fmt.Errorf("failed to execute operation: %w", err)
	}

	// Check if tool execution was successful
	if !result.Success {
		return ChatResponse{
			Reply:                fmt.Sprintf("Operation failed: %s", result.Message),
			Intent:               IntentUnknown,
			RequiresConfirmation: false,
		}, nil
	}

	// Return success response
	response := ChatResponse{
		Reply:                result.Message,
		Intent:               IntentUnknown,
		Data:                 result.Data,
		RequiresConfirmation: false,
	}

	log.Printf("💬 Confirmation: Operation %s completed successfully", toolName)
	return response, nil
}

// ============================================================================
// Legacy Helper Methods (DEPRECATED - kept for reference)
// These methods are no longer used with MCP integration but kept for
// backward compatibility and reference purposes.
// ============================================================================

// handleBalanceIntent handles balance queries
func (s *Service) handleBalanceIntent(userID string) (ChatResponse, error) {
	balance, err := s.accountService.GetBalance(userID)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("failed to retrieve balance: %w", err)
	}

	// Convert cents to dollars for display
	dollars := float64(balance) / 100.0

	return ChatResponse{
		Reply:  fmt.Sprintf("Your current balance is $%.2f", dollars),
		Intent: IntentBalance,
		Data: map[string]interface{}{
			"balance":       balance,
			"balance_usd":   dollars,
			"currency":      "USD",
		},
		RequiresConfirmation: false,
	}, nil
}

// handleDepositIntent handles deposit requests (requires confirmation)
func (s *Service) handleDepositIntent(amount int64) ChatResponse {
	dollars := float64(amount) / 100.0

	return ChatResponse{
		Reply:  fmt.Sprintf("You want to deposit $%.2f to your account. Please confirm to proceed.", dollars),
		Intent: IntentDeposit,
		Data: map[string]interface{}{
			"amount":     amount,
			"amount_usd": dollars,
			"action":     "deposit",
		},
		RequiresConfirmation: true,
	}
}

// handleWithdrawIntent handles withdrawal requests (requires confirmation)
func (s *Service) handleWithdrawIntent(amount int64) ChatResponse {
	dollars := float64(amount) / 100.0

	return ChatResponse{
		Reply:  fmt.Sprintf("You want to withdraw $%.2f from your account. Please confirm to proceed.", dollars),
		Intent: IntentWithdraw,
		Data: map[string]interface{}{
			"amount":     amount,
			"amount_usd": dollars,
			"action":     "withdraw",
		},
		RequiresConfirmation: true,
	}
}

// handleTransferIntent handles transfer requests (requires confirmation)
func (s *Service) handleTransferIntent(amount int64, toAccountID uint64) ChatResponse {
	dollars := float64(amount) / 100.0

	return ChatResponse{
		Reply: fmt.Sprintf("You want to transfer $%.2f to account %d. Please confirm to proceed.", dollars, toAccountID),
		Intent: IntentTransfer,
		Data: map[string]interface{}{
			"amount":        amount,
			"amount_usd":    dollars,
			"to_account_id": toAccountID,
			"action":        "transfer",
		},
		RequiresConfirmation: true,
	}
}

// handleHistoryIntent handles transaction history queries
func (s *Service) handleHistoryIntent(userID string, limit int) (ChatResponse, error) {
	// Default to page 1, use the limit from the parsed intent
	history, err := s.transactionService.GetHistory(userID, 1, limit)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("failed to retrieve transaction history: %w", err)
	}

	count := len(history)
	reply := fmt.Sprintf("Here are your last %d transactions.", count)

	if count == 0 {
		reply = "You don't have any transactions yet."
	}

	return ChatResponse{
		Reply:  reply,
		Intent: IntentHistory,
		Data: map[string]interface{}{
			"transactions": history,
			"count":        count,
			"limit":        limit,
		},
		RequiresConfirmation: false,
	}, nil
}

// buildUnknownResponse builds a response for unrecognized intents
func (s *Service) buildUnknownResponse(customMessage string) ChatResponse {
	reply := "I didn't understand that request. Here are some things you can ask me:"

	if customMessage != "" {
		reply = customMessage + "\n\nHere are some things you can ask me:"
	}

	reply += "\n\n" +
		"- \"What's my balance?\"\n" +
		"- \"Deposit $100\"\n" +
		"- \"Withdraw $50\"\n" +
		"- \"Transfer $200 to account 12345\"\n" +
		"- \"Show my last 10 transactions\""

	return ChatResponse{
		Reply:  reply,
		Intent: IntentUnknown,
		Data: map[string]interface{}{
			"suggestions": []string{
				"What's my balance?",
				"Deposit $100",
				"Withdraw $50",
				"Transfer $200 to account 12345",
				"Show my last 10 transactions",
			},
		},
		RequiresConfirmation: false,
	}
}
