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
	repo     *Repository
}

// NewService creates a new transaction service
func NewService(db *gorm.DB, tbClient *tigerbeetle.Client) *Service {
	return &Service{
		db:       db,
		tbClient: tbClient,
		repo:     NewRepository(db),
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

	// Generate transfer ID
	transferID := tb_types.ToUint128(uint64(uuid.New().ID()))

	// Create transfer from system account to user account
	transfers := []tb_types.Transfer{
		{
			ID:              transferID,
			DebitAccountID:  s.tbClient.SystemAccountID,        // System account (source)
			CreditAccountID: tb_types.ToUint128(user.TigerBeetleAccountID), // User account (destination)
			Amount:          tb_types.ToUint128(uint64(amount)),
			Ledger:          1,
			Code:            1, // Deposit code
		},
	}

	// Execute transfer in TigerBeetle
	results, err := s.tbClient.CreateTransfers(transfers)
	if err != nil {
		return fmt.Errorf("failed to create transfer: %w", err)
	}

	// Check for errors
	if len(results) > 0 {
		return fmt.Errorf("transfer failed with result code: %d", results[0].Result)
	}

	// Create transaction record in PostgreSQL (audit log)
	// Convert SystemAccountID (Uint128) to uint64 for PostgreSQL storage
	systemAcctBI := s.tbClient.SystemAccountID.BigInt()

	txRecord := &models.Transaction{
		UserID:          user.ID,
		Type:            models.TransactionTypeDeposit,
		Amount:          amount,
		DebitAccountID:  systemAcctBI.Uint64(),
		CreditAccountID: user.TigerBeetleAccountID,
		Status:          models.TransactionStatusCompleted,
		Description:     fmt.Sprintf("Deposit of %d cents", amount),
	}
	txRecord.SetTigerBeetleTransferID(transferID)

	// Save to PostgreSQL - non-blocking (TigerBeetle is source of truth)
	if err := s.repo.Create(txRecord); err != nil {
		// Log error but don't fail the request (money already transferred in TigerBeetle)
		log.Printf("🚨 CRITICAL: Deposit transfer %s succeeded in TigerBeetle but failed to log in PostgreSQL: %v",
			txRecord.TigerBeetleTransferID, err)
		log.Printf("   UserID: %s, Amount: %d, TigerBeetle Account: %d", userID, amount, user.TigerBeetleAccountID)
		// Continue execution - the deposit succeeded in TigerBeetle
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

	// Generate transfer ID
	transferID := tb_types.ToUint128(uint64(uuid.New().ID()))

	// Create transfer from user account to system account
	transfers := []tb_types.Transfer{
		{
			ID:              transferID,
			DebitAccountID:  tb_types.ToUint128(user.TigerBeetleAccountID), // User account (source)
			CreditAccountID: s.tbClient.SystemAccountID,        // System account (destination)
			Amount:          tb_types.ToUint128(uint64(amount)),
			Ledger:          1,
			Code:            2, // Withdrawal code
		},
	}

	// Execute transfer in TigerBeetle
	results, err := s.tbClient.CreateTransfers(transfers)
	if err != nil {
		return fmt.Errorf("failed to create transfer: %w", err)
	}

	// Check for errors
	if len(results) > 0 {
		return fmt.Errorf("transfer failed with result code: %d", results[0].Result)
	}

	// Create transaction record in PostgreSQL (audit log)
	// Convert SystemAccountID (Uint128) to uint64 for PostgreSQL storage
	systemAcctBI := s.tbClient.SystemAccountID.BigInt()

	txRecord := &models.Transaction{
		UserID:          user.ID,
		Type:            models.TransactionTypeWithdraw,
		Amount:          amount,
		DebitAccountID:  user.TigerBeetleAccountID,
		CreditAccountID: systemAcctBI.Uint64(),
		Status:          models.TransactionStatusCompleted,
		Description:     fmt.Sprintf("Withdrawal of %d cents", amount),
	}
	txRecord.SetTigerBeetleTransferID(transferID)

	// Save to PostgreSQL - non-blocking
	if err := s.repo.Create(txRecord); err != nil {
		log.Printf("🚨 CRITICAL: Withdrawal transfer %s succeeded in TigerBeetle but failed to log in PostgreSQL: %v",
			txRecord.TigerBeetleTransferID, err)
		log.Printf("   UserID: %s, Amount: %d, TigerBeetle Account: %d", userID, amount, user.TigerBeetleAccountID)
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
	log.Printf("🔍 [Transfer] Verifying destination account exists in TigerBeetle (AccountID: %d)...", toAccountID)
	destAccounts, err := s.tbClient.LookupAccounts([]tb_types.Uint128{tb_types.ToUint128(toAccountID)})
	if err != nil || len(destAccounts) == 0 {
		log.Printf("❌ [Transfer] Destination account %d not found in TigerBeetle", toAccountID)
		return fmt.Errorf("destination account not found")
	}
	log.Printf("✅ [Transfer] Destination account %d exists in TigerBeetle", toAccountID)

	// Find recipient user by TigerBeetle account ID
	log.Printf("🔍 [Transfer] Searching for recipient user in PostgreSQL by tigerbeetle_account_id = %d...", toAccountID)
	var toUser models.User
	if err := s.db.Where("tigerbeetle_account_id = ?", toAccountID).First(&toUser).Error; err != nil {
		// Recipient not found in PostgreSQL (might be system account or deleted user)
		log.Printf("❌ [Transfer] Recipient with TigerBeetle account %d NOT FOUND in PostgreSQL: %v", toAccountID, err)
		log.Printf("❌ [Transfer] toUser.ID will be uuid.Nil, RecipientUserID will NOT be set")
	} else {
		log.Printf("✅ [Transfer] Recipient user FOUND: ID=%s, Email=%s, TigerBeetleAccountID=%d", toUser.ID, toUser.Email, toUser.TigerBeetleAccountID)
	}

	// Generate transfer ID
	transferID := tb_types.ToUint128(uint64(uuid.New().ID()))

	// Create transfer between user accounts
	transfers := []tb_types.Transfer{
		{
			ID:              transferID,
			DebitAccountID:  tb_types.ToUint128(fromUser.TigerBeetleAccountID), // Sender
			CreditAccountID: tb_types.ToUint128(toAccountID),                   // Recipient
			Amount:          tb_types.ToUint128(uint64(amount)),
			Ledger:          1,
			Code:            3, // Transfer code
		},
	}

	// Execute transfer in TigerBeetle
	results, err := s.tbClient.CreateTransfers(transfers)
	if err != nil {
		return fmt.Errorf("failed to create transfer: %w", err)
	}

	// Check for errors
	if len(results) > 0 {
		return fmt.Errorf("transfer failed with result code: %d", results[0].Result)
	}

	// Create transaction record in PostgreSQL (audit log)
	log.Printf("🔍 [Transfer] Creating transaction record for PostgreSQL...")
	txRecord := &models.Transaction{
		UserID:          fromUser.ID,
		RecipientUserID: nil, // Will set if recipient found
		Type:            models.TransactionTypeTransfer,
		Amount:          amount,
		DebitAccountID:  fromUser.TigerBeetleAccountID,
		CreditAccountID: toAccountID,
		Status:          models.TransactionStatusCompleted,
		Description:     fmt.Sprintf("Transfer of %d cents to account %d", amount, toAccountID),
	}
	txRecord.SetTigerBeetleTransferID(transferID)

	// Set recipient user ID if found
	log.Printf("🔍 [Transfer] Checking if toUser.ID != uuid.Nil (toUser.ID = %s)...", toUser.ID)
	if toUser.ID != uuid.Nil {
		txRecord.RecipientUserID = &toUser.ID
		log.Printf("✅ [Transfer] RecipientUserID SET to: %s", toUser.ID)
	} else {
		log.Printf("❌ [Transfer] RecipientUserID NOT SET (toUser.ID is uuid.Nil)")
	}

	log.Printf("🔍 [Transfer] Transaction record to save:")
	log.Printf("   UserID: %s", txRecord.UserID)
	log.Printf("   RecipientUserID: %v", txRecord.RecipientUserID)
	log.Printf("   Type: %s", txRecord.Type)
	log.Printf("   Amount: %d cents", txRecord.Amount)
	log.Printf("   DebitAccountID: %d", txRecord.DebitAccountID)
	log.Printf("   CreditAccountID: %d", txRecord.CreditAccountID)
	log.Printf("   Status: %s", txRecord.Status)

	// Save to PostgreSQL - non-blocking
	if err := s.repo.Create(txRecord); err != nil {
		log.Printf("🚨 CRITICAL: Transfer %s succeeded in TigerBeetle but failed to log in PostgreSQL: %v",
			txRecord.TigerBeetleTransferID, err)
		log.Printf("   From UserID: %s, To Account: %d, Amount: %d", fromUserID, toAccountID, amount)
	}

	log.Printf("✅ Transfer successful: %d cents from user %s to account %d", amount, fromUserID, toAccountID)
	return nil
}

// GetHistory retrieves transaction history for a user from PostgreSQL
// Returns paginated transaction records with enriched user information
func (s *Service) GetHistory(userID string, page, limit int) ([]models.TransactionDTO, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get user to validate existence
	user, err := s.getUserByID(userID)
	if err != nil {
		return nil, err
	}

	log.Printf("📋 [GetHistory] Retrieving transaction history for user %s (UUID: %s, page %d, limit %d)", userID, user.ID, page, limit)

	// Get transactions from repository (includes all transactions where user is sender or recipient)
	log.Printf("📋 [GetHistory] Calling repository.GetAllByUserID...")
	transactions, err := s.repo.GetAllByUserID(user.ID, page, limit)
	if err != nil {
		log.Printf("❌ [GetHistory] Failed to retrieve transactions: %v", err)
		return nil, fmt.Errorf("failed to retrieve transaction history: %w", err)
	}
	log.Printf("📋 [GetHistory] Repository returned %d raw transactions", len(transactions))

	// Log each transaction for debugging
	for i, tx := range transactions {
		log.Printf("📋 [GetHistory] Transaction %d: ID=%s, UserID=%s, RecipientUserID=%v, Type=%s, Amount=%d",
			i+1, tx.ID, tx.UserID, tx.RecipientUserID, tx.Type, tx.Amount)
	}

	// Convert to DTOs with enriched information
	dtos := make([]models.TransactionDTO, 0, len(transactions))
	for _, tx := range transactions {
		dtos = append(dtos, tx.ToDTO())
	}

	log.Printf("✅ [GetHistory] Retrieved %d transactions for user %s", len(dtos), userID)
	return dtos, nil
}

// GetHistoryCount returns the total count of transactions for pagination
func (s *Service) GetHistoryCount(userID string) (int64, error) {
	user, err := s.getUserByID(userID)
	if err != nil {
		return 0, err
	}

	return s.repo.CountAllByUserID(user.ID)
}
