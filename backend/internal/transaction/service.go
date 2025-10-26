package transaction

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/hlabs/banking-system/internal/models"
	"github.com/hlabs/banking-system/internal/tigerbeetle"
	tb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"gorm.io/gorm"
)

// Service handles transaction-related business logic
type Service struct {
	db       *gorm.DB
	tbClient *tigerbeetle.Client
}

// NewService creates a new transaction service
func NewService(db *gorm.DB, tbClient *tigerbeetle.Client) *Service {
	return &Service{
		db:       db,
		tbClient: tbClient,
	}
}

// getUserByID retrieves a user by their ID (helper method)
func (s *Service) getUserByID(userID string) (*models.User, error) {
	var user models.User

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	if err := s.db.First(&user, "id = ?", uid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}

// Deposit adds funds to a user's account (from system account)
func (s *Service) Deposit(userID string, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("deposit amount must be positive")
	}

	// Get user to retrieve TigerBeetle account ID
	user, err := s.getUserByID(userID)
	if err != nil {
		return err
	}

	// Create transfer from system account to user account
	transfers := []tb_types.Transfer{
		{
			ID:              tb_types.ToUint128(uint64(uuid.New().ID())),
			DebitAccountID:  s.tbClient.SystemAccountID,        // System account (source)
			CreditAccountID: tb_types.ToUint128(user.TigerBeetleAccountID), // User account (destination)
			Amount:          tb_types.ToUint128(uint64(amount)),
			Ledger:          1,
			Code:            1, // Deposit code
		},
	}

	// Execute transfer
	results, err := s.tbClient.CreateTransfers(transfers)
	if err != nil {
		return fmt.Errorf("failed to create transfer: %w", err)
	}

	// Check for errors
	if len(results) > 0 {
		return fmt.Errorf("transfer failed with result code: %d", results[0].Result)
	}

	log.Printf("✅ Deposit successful: %d cents to user %s (TB Account: %d)", amount, userID, user.TigerBeetleAccountID)
	return nil
}

// Withdraw removes funds from a user's account (to system account)
func (s *Service) Withdraw(userID string, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("withdrawal amount must be positive")
	}

	// Get user
	user, err := s.getUserByID(userID)
	if err != nil {
		return err
	}

	// Check balance first
	balance, err := s.tbClient.GetBalance(user.TigerBeetleAccountID)
	if err != nil {
		return fmt.Errorf("failed to check balance: %w", err)
	}

	if balance < amount {
		return fmt.Errorf("insufficient funds: balance is %d, requested %d", balance, amount)
	}

	// Create transfer from user account to system account
	transfers := []tb_types.Transfer{
		{
			ID:              tb_types.ToUint128(uint64(uuid.New().ID())),
			DebitAccountID:  tb_types.ToUint128(user.TigerBeetleAccountID), // User account (source)
			CreditAccountID: s.tbClient.SystemAccountID,        // System account (destination)
			Amount:          tb_types.ToUint128(uint64(amount)),
			Ledger:          1,
			Code:            2, // Withdrawal code
		},
	}

	// Execute transfer
	results, err := s.tbClient.CreateTransfers(transfers)
	if err != nil {
		return fmt.Errorf("failed to create transfer: %w", err)
	}

	// Check for errors
	if len(results) > 0 {
		return fmt.Errorf("transfer failed with result code: %d", results[0].Result)
	}

	log.Printf("✅ Withdrawal successful: %d cents from user %s (TB Account: %d)", amount, userID, user.TigerBeetleAccountID)
	return nil
}

// Transfer sends funds from one user to another
func (s *Service) Transfer(fromUserID string, toAccountID uint64, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("transfer amount must be positive")
	}

	// Get sender
	fromUser, err := s.getUserByID(fromUserID)
	if err != nil {
		return fmt.Errorf("sender not found: %w", err)
	}

	// Check sender's balance
	balance, err := s.tbClient.GetBalance(fromUser.TigerBeetleAccountID)
	if err != nil {
		return fmt.Errorf("failed to check balance: %w", err)
	}

	if balance < amount {
		return fmt.Errorf("insufficient funds: balance is %d, requested %d", balance, amount)
	}

	// Verify destination account exists (lookup in TigerBeetle)
	destAccounts, err := s.tbClient.LookupAccounts([]tb_types.Uint128{tb_types.ToUint128(toAccountID)})
	if err != nil || len(destAccounts) == 0 {
		return fmt.Errorf("destination account not found")
	}

	// Create transfer between user accounts
	transfers := []tb_types.Transfer{
		{
			ID:              tb_types.ToUint128(uint64(uuid.New().ID())),
			DebitAccountID:  tb_types.ToUint128(fromUser.TigerBeetleAccountID), // Sender
			CreditAccountID: tb_types.ToUint128(toAccountID),                   // Recipient
			Amount:          tb_types.ToUint128(uint64(amount)),
			Ledger:          1,
			Code:            3, // Transfer code
		},
	}

	// Execute transfer
	results, err := s.tbClient.CreateTransfers(transfers)
	if err != nil {
		return fmt.Errorf("failed to create transfer: %w", err)
	}

	// Check for errors
	if len(results) > 0 {
		return fmt.Errorf("transfer failed with result code: %d", results[0].Result)
	}

	log.Printf("✅ Transfer successful: %d cents from user %s to account %d", amount, fromUserID, toAccountID)
	return nil
}

// GetHistory retrieves transaction history for a user
// Note: This is a simplified version. TigerBeetle doesn't have built-in transaction history queries,
// so you would typically need to store this in PostgreSQL or query TigerBeetle transfers
func (s *Service) GetHistory(userID string, page, limit int) ([]map[string]interface{}, error) {
	// Get user
	user, err := s.getUserByID(userID)
	if err != nil {
		return nil, err
	}

	// For now, return a placeholder
	// In a real implementation, you would:
	// 1. Store transaction records in PostgreSQL when transfers are created
	// 2. Or query TigerBeetle's transfer log (more complex)

	log.Printf("📋 Retrieving transaction history for user %s (TB Account: %d)", userID, user.TigerBeetleAccountID)

	// Placeholder response
	history := []map[string]interface{}{
		{
			"id":        "example-1",
			"type":      "deposit",
			"amount":    10000,
			"timestamp": "2024-10-26T00:00:00Z",
			"status":    "completed",
		},
	}

	return history, nil
}
