package chat

import (
	"fmt"
	"log"

	"github.com/hlabs/banking-system/internal/account"
	"github.com/hlabs/banking-system/internal/transaction"
)

// Service handles chat-related business logic
type Service struct {
	accountService     *account.Service
	transactionService *transaction.Service
}

// NewService creates a new chat service
func NewService(accountService *account.Service, transactionService *transaction.Service) *Service {
	return &Service{
		accountService:     accountService,
		transactionService: transactionService,
	}
}

// ProcessMessage processes a user's chat message and returns an appropriate response
func (s *Service) ProcessMessage(userID, message string) (ChatResponse, error) {
	// Log chat interaction for audit
	log.Printf("💬 Chat: User %s: %s", userID, message)

	// Parse intent from message
	parsed := ParseIntent(message)

	// Validate intent has required parameters
	if err := ValidateIntent(parsed); err != nil {
		return s.buildUnknownResponse(err.Error()), nil
	}

	// Execute intent
	var response ChatResponse
	var err error

	switch parsed.Intent {
	case IntentBalance:
		response, err = s.handleBalanceIntent(userID)

	case IntentDeposit:
		response = s.handleDepositIntent(parsed.Amount)

	case IntentWithdraw:
		response = s.handleWithdrawIntent(parsed.Amount)

	case IntentTransfer:
		response = s.handleTransferIntent(parsed.Amount, parsed.ToAccountID)

	case IntentHistory:
		response, err = s.handleHistoryIntent(userID, parsed.Limit)

	default:
		response = s.buildUnknownResponse("")
	}

	if err != nil {
		return ChatResponse{}, err
	}

	log.Printf("💬 Chat: Response - Intent: %s, Confirmation: %v", response.Intent, response.RequiresConfirmation)
	return response, nil
}

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
