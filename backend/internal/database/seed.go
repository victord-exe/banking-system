package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/hlabs/banking-system/internal/models"
	"github.com/hlabs/banking-system/internal/tigerbeetle"
	"github.com/hlabs/banking-system/pkg/utils"
	tb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TestUser represents a user from the test data JSON file
type TestUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"` // Plaintext password from JSON
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

// TestAccount represents an account from the test data JSON file
type TestAccount struct {
	AccountNumber  string  `json:"account_number"`
	UserID         string  `json:"user_id"`
	InitialBalance float64 `json:"initial_balance"`
	Currency       string  `json:"currency"`
	AccountType    string  `json:"account_type"`
}

// TestTransaction represents a transaction from the test data JSON file
type TestTransaction struct {
	FromAccount string    `json:"from_account"`
	ToAccount   string    `json:"to_account"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
}

// TestDataFile represents the structure of the test data JSON
type TestDataFile struct {
	Users        []TestUser        `json:"users"`
	Accounts     []TestAccount     `json:"accounts"`
	Transactions []TestTransaction `json:"transactions"`
}

// Seed populates the database with test users, accounts, and transactions from datos-prueba-HNL.json
// This function is idempotent - it only seeds if the users table is empty
func Seed(db *gorm.DB, tbClient *tigerbeetle.Client) error {
	log.Println("================================================================")
	log.Println("🌱 DATABASE SEEDING - Starting initialization...")
	log.Println("================================================================")

	// Check if users already exist
	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	if count > 0 {
		log.Printf("ℹ️  Database already contains %d users, skipping seed", count)
		log.Println("================================================================")
		return nil
	}

	// Path to test data file (mounted by docker-compose)
	testDataPath := "/app/datos-prueba-HNL.json"

	// Check if file exists
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		// Not a fatal error - just log and continue
		log.Printf("⚠️  Test data file not found at %s, skipping seed", testDataPath)
		log.Println("================================================================")
		return nil
	}

	// Read test data file
	log.Printf("📖 Reading test data from: %s", testDataPath)
	fileData, err := os.ReadFile(testDataPath)
	if err != nil {
		return fmt.Errorf("failed to read test data file: %w", err)
	}

	// Validate UTF-8 encoding
	if !utf8.Valid(fileData) {
		log.Printf("⚠️  WARNING: Test data file contains invalid UTF-8 sequences - attempting to parse anyway")
		// Note: We continue because json.Unmarshal can often handle this,
		// but we want to alert the user to potential encoding issues
	} else {
		log.Printf("✅ UTF-8 validation passed - file encoding is correct")
	}

	// Parse JSON
	log.Println("📋 Parsing JSON data...")
	var testData TestDataFile
	if err := json.Unmarshal(fileData, &testData); err != nil {
		return fmt.Errorf("failed to parse test data JSON: %w", err)
	}

	log.Printf("📊 Data summary:")
	log.Printf("   Users: %d", len(testData.Users))
	log.Printf("   Accounts: %d", len(testData.Accounts))
	log.Printf("   Transactions: %d", len(testData.Transactions))
	log.Println("================================================================")

	// PHASE 1: Create users
	if err := seedUsers(db, tbClient, testData.Users); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	// PHASE 2: Link accounts and set initial balances
	if err := seedAccounts(db, tbClient, testData.Accounts); err != nil {
		return fmt.Errorf("failed to seed accounts: %w", err)
	}

	// PHASE 3: Load historical transactions
	if err := seedTransactions(db, tbClient, testData.Transactions); err != nil {
		return fmt.Errorf("failed to seed transactions: %w", err)
	}

	log.Println("================================================================")
	log.Println("✅ DATABASE SEEDING COMPLETED SUCCESSFULLY")
	log.Println("================================================================")

	// DIAGNOSTIC: Show statistics
	if err := showSeedingStatistics(db, tbClient); err != nil {
		log.Printf("⚠️  Failed to show statistics: %v", err)
	}

	return nil
}

// seedUsers creates users in PostgreSQL and TigerBeetle
func seedUsers(db *gorm.DB, tbClient *tigerbeetle.Client, users []TestUser) error {
	log.Println("================================================================")
	log.Println("👥 PHASE 1: Creating Users")
	log.Println("================================================================")

	totalUsers := len(users)
	successCount := 0
	failCount := 0
	startTime := time.Now()

	for i, testUser := range users {
		// Progress indicator
		progress := float64(i+1) / float64(totalUsers) * 100
		showProgress := (i+1)%500 == 0 || i == 0 || i == totalUsers-1

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testUser.Password), bcrypt.DefaultCost)
		if err != nil {
			if showProgress {
				log.Printf("❌ [%d/%d] %.1f%% - Failed: %s (password hash error)", i+1, totalUsers, progress, testUser.Email)
			}
			failCount++
			continue
		}

		// Generate unique TigerBeetle account ID
		tbAccountID := utils.GenerateAccountID()

		// Create TigerBeetle account (empty initially, balance set in Phase 2)
		if err := tbClient.CreateAccount(tbAccountID); err != nil {
			if showProgress {
				log.Printf("❌ [%d/%d] %.1f%% - Failed: %s (TigerBeetle error)", i+1, totalUsers, progress, testUser.Email)
			}
			failCount++
			continue
		}

		// Parse UUID from test data
		userUUID, err := uuid.Parse(testUser.ID)
		if err != nil {
			// Silently generate new UUID if invalid
			userUUID = uuid.New()
		}

		// Create user in PostgreSQL
		user := models.User{
			ID:                   userUUID,
			Email:                testUser.Email,
			Password:             string(hashedPassword),
			FullName:             testUser.FullName,
			TigerBeetleAccountID: tbAccountID,
			CreatedAt:            testUser.CreatedAt,
		}

		if err := db.Create(&user).Error; err != nil {
			if showProgress {
				log.Printf("❌ [%d/%d] %.1f%% - Failed: %s (database error)", i+1, totalUsers, progress, testUser.Email)
			}
			failCount++
			continue
		}

		successCount++

		// Show progress only at milestones
		if showProgress {
			log.Printf("✅ [%d/%d] %.1f%% - Created users...", i+1, totalUsers, progress)
		}
	}

	duration := time.Since(startTime)

	log.Println("----------------------------------------------------------------")
	log.Printf("✅ Phase 1 Complete - Users Created: %d/%d (⏱️  %v)", successCount, totalUsers, duration)
	if failCount > 0 {
		log.Printf("⚠️  Failed: %d users", failCount)
	}
	log.Println("================================================================")

	if successCount == 0 {
		return fmt.Errorf("all user creations failed")
	}

	return nil
}

// seedAccounts links account numbers to users and sets initial balances
func seedAccounts(db *gorm.DB, tbClient *tigerbeetle.Client, accounts []TestAccount) error {
	log.Println("================================================================")
	log.Println("💳 PHASE 2: Linking Accounts & Setting Initial Balances")
	log.Println("================================================================")

	totalAccounts := len(accounts)
	successCount := 0
	failCount := 0
	startTime := time.Now()

	// Create map of user_id -> TigerBeetle account ID for quick lookup
	userMap := make(map[string]uint64)
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}

	for _, user := range users {
		userMap[user.ID.String()] = user.TigerBeetleAccountID
	}

	log.Printf("📋 Loaded %d users into lookup map", len(userMap))

	for i, account := range accounts {
		progress := float64(i+1) / float64(totalAccounts) * 100
		showProgress := (i+1)%500 == 0 || i == 0 || i == totalAccounts-1

		// Find user's TigerBeetle account ID
		tbAccountID, exists := userMap[account.UserID]
		if !exists {
			if showProgress {
				log.Printf("⚠️  [%d/%d] %.1f%% - Skipping: %s (user not found)", i+1, totalAccounts, progress, account.AccountNumber)
			}
			failCount++
			continue
		}

		// Update user with account number
		if err := db.Model(&models.User{}).
			Where("id = ?", account.UserID).
			Update("account_number", account.AccountNumber).Error; err != nil {
			if showProgress {
				log.Printf("❌ [%d/%d] %.1f%% - Failed to update account number: %s", i+1, totalAccounts, progress, account.AccountNumber)
			}
			failCount++
			continue
		}

		// Set initial balance via deposit from system account
		// Convert dollars to cents
		amountCents := int64(account.InitialBalance * 100)

		if amountCents > 0 {
			// Generate transfer ID
			transferID := tb_types.ToUint128(uint64(uuid.New().ID()))

			// Create transfer from system account to user account
			transfers := []tb_types.Transfer{
				{
					ID:              transferID,
					DebitAccountID:  tbClient.SystemAccountID,                 // System account (source)
					CreditAccountID: tb_types.ToUint128(tbAccountID),          // User account (destination)
					Amount:          tb_types.ToUint128(uint64(amountCents)),
					Ledger:          1,
					Code:            100, // Initial balance deposit code
				},
			}

			// Execute transfer in TigerBeetle
			results, err := tbClient.CreateTransfers(transfers)
			if err != nil || len(results) > 0 {
				if showProgress {
					log.Printf("❌ [%d/%d] %.1f%% - Failed to set balance: %s", i+1, totalAccounts, progress, account.AccountNumber)
				}
				failCount++
				continue
			}

			// Record transaction in PostgreSQL
			userUUID, _ := uuid.Parse(account.UserID)
			systemAcctBI := tbClient.SystemAccountID.BigInt()

			txRecord := &models.Transaction{
				UserID:          userUUID,
				Type:            models.TransactionTypeDeposit,
				Amount:          amountCents,
				DebitAccountID:  systemAcctBI.Uint64(),
				CreditAccountID: tbAccountID,
				Status:          models.TransactionStatusCompleted,
				Description:     fmt.Sprintf("Initial balance: $%.2f", account.InitialBalance),
			}
			txRecord.SetTigerBeetleTransferID(transferID)

			// Save to PostgreSQL (non-critical if it fails)
			if err := db.Create(txRecord).Error; err != nil {
				log.Printf("⚠️  Failed to log initial balance transaction for %s", account.AccountNumber)
			}
		}

		successCount++

		if showProgress {
			log.Printf("✅ [%d/%d] %.1f%% - Linked accounts and set balances...", i+1, totalAccounts, progress)
		}
	}

	duration := time.Since(startTime)

	log.Println("----------------------------------------------------------------")
	log.Printf("✅ Phase 2 Complete - Accounts Linked: %d/%d (⏱️  %v)", successCount, totalAccounts, duration)
	if failCount > 0 {
		log.Printf("⚠️  Failed: %d accounts", failCount)
	}
	log.Println("================================================================")

	return nil
}

// seedTransactions loads historical transactions from the test data
func seedTransactions(db *gorm.DB, tbClient *tigerbeetle.Client, transactions []TestTransaction) error {
	log.Println("================================================================")
	log.Println("💸 PHASE 3: Loading Historical Transactions")
	log.Println("================================================================")

	// Sort transactions by timestamp (chronological order)
	log.Println("📋 Sorting transactions chronologically...")
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Timestamp.Before(transactions[j].Timestamp)
	})

	totalTransactions := len(transactions)
	successCount := 0
	failCount := 0
	skippedCount := 0
	startTime := time.Now()

	// Create map of account_number -> user for quick lookup
	accountMap := make(map[string]*models.User)
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}

	for i := range users {
		if users[i].AccountNumber != "" {
			accountMap[users[i].AccountNumber] = &users[i]
		}
	}

	log.Printf("📋 Loaded %d accounts into lookup map", len(accountMap))
	log.Println("🚀 Processing transactions...")

	for i, tx := range transactions {
		progress := float64(i+1) / float64(totalTransactions) * 100
		showProgress := (i+1)%1000 == 0 || i == 0 || i == totalTransactions-1

		// Skip if not completed
		if tx.Status != "completed" {
			skippedCount++
			continue
		}

		// Convert amount to cents
		amountCents := int64(tx.Amount * 100)
		if amountCents <= 0 {
			skippedCount++
			continue
		}

		// Determine debit and credit accounts
		var debitTBAccountID, creditTBAccountID uint64
		var fromUser, toUser *models.User

		// Handle from_account
		if tx.FromAccount == "EXTERNAL" {
			// External source = system account
			systemAcctBI := tbClient.SystemAccountID.BigInt()
			debitTBAccountID = systemAcctBI.Uint64()
		} else {
			fromUser = accountMap[tx.FromAccount]
			if fromUser == nil {
				skippedCount++
				continue
			}
			debitTBAccountID = fromUser.TigerBeetleAccountID
		}

		// Handle to_account
		if tx.ToAccount == "EXTERNAL" {
			// External destination = system account
			systemAcctBI := tbClient.SystemAccountID.BigInt()
			creditTBAccountID = systemAcctBI.Uint64()
		} else {
			toUser = accountMap[tx.ToAccount]
			if toUser == nil {
				skippedCount++
				continue
			}
			creditTBAccountID = toUser.TigerBeetleAccountID
		}

		// Generate transfer ID
		transferID := tb_types.ToUint128(uint64(uuid.New().ID()))

		// Create transfer in TigerBeetle
		transfers := []tb_types.Transfer{
			{
				ID:              transferID,
				DebitAccountID:  tb_types.ToUint128(debitTBAccountID),
				CreditAccountID: tb_types.ToUint128(creditTBAccountID),
				Amount:          tb_types.ToUint128(uint64(amountCents)),
				Ledger:          1,
				Code:            4, // Historical transaction code
			},
		}

		results, err := tbClient.CreateTransfers(transfers)
		if err != nil || len(results) > 0 {
			// Transfer failed (likely insufficient funds or other constraint)
			failCount++
			continue
		}

		// Record in PostgreSQL (use the user who sent/received the money)
		var primaryUser *models.User
		var recipientUserID *uuid.UUID
		var txType models.TransactionType

		if fromUser != nil {
			primaryUser = fromUser
			txType = models.TransactionTypeTransfer
			if toUser != nil {
				recipientUserID = &toUser.ID
			}
		} else if toUser != nil {
			primaryUser = toUser
			txType = models.TransactionTypeDeposit
		}

		if primaryUser != nil {
			txRecord := &models.Transaction{
				UserID:          primaryUser.ID,
				RecipientUserID: recipientUserID,
				Type:            txType,
				Amount:          amountCents,
				DebitAccountID:  debitTBAccountID,
				CreditAccountID: creditTBAccountID,
				Status:          models.TransactionStatusCompleted,
				Description:     tx.Description,
				CreatedAt:       tx.Timestamp,
			}
			txRecord.SetTigerBeetleTransferID(transferID)

			if err := db.Create(txRecord).Error; err != nil {
				// Log but don't fail - TigerBeetle is source of truth
				log.Printf("⚠️  Failed to log transaction in PostgreSQL: %v", err)
			}
		}

		successCount++

		if showProgress {
			log.Printf("✅ [%d/%d] %.1f%% - Processed transactions...", i+1, totalTransactions, progress)
		}
	}

	duration := time.Since(startTime)

	log.Println("----------------------------------------------------------------")
	log.Printf("✅ Phase 3 Complete - Transactions Loaded: %d/%d (⏱️  %v)", successCount, totalTransactions, duration)
	if skippedCount > 0 {
		log.Printf("⚠️  Skipped: %d transactions (incomplete or invalid)", skippedCount)
	}
	if failCount > 0 {
		log.Printf("❌ Failed: %d transactions", failCount)
	}
	log.Println("================================================================")

	return nil
}

// showSeedingStatistics displays diagnostic information after seeding
func showSeedingStatistics(db *gorm.DB, tbClient *tigerbeetle.Client) error {
	log.Println("")
	log.Println("================================================================")
	log.Println("📊 SEEDING STATISTICS & DIAGNOSTICS")
	log.Println("================================================================")

	// Count users
	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}
	log.Printf("👥 Total Users: %d", userCount)

	// Count users with account numbers
	var usersWithAccounts int64
	if err := db.Model(&models.User{}).Where("account_number IS NOT NULL AND account_number != ''").Count(&usersWithAccounts).Error; err != nil {
		return fmt.Errorf("failed to count users with accounts: %w", err)
	}
	log.Printf("💳 Users with Account Numbers: %d", usersWithAccounts)

	// Count transactions
	var txCount int64
	if err := db.Model(&models.Transaction{}).Count(&txCount).Error; err != nil {
		return fmt.Errorf("failed to count transactions: %w", err)
	}
	log.Printf("💸 Total Transactions: %d", txCount)

	// Get sample users with balances
	log.Println("----------------------------------------------------------------")
	log.Println("📋 Sample Account Balances (from TigerBeetle):")
	log.Println("----------------------------------------------------------------")

	var sampleUsers []models.User
	if err := db.Where("account_number IS NOT NULL AND account_number != ''").
		Limit(5).
		Find(&sampleUsers).Error; err != nil {
		return fmt.Errorf("failed to get sample users: %w", err)
	}

	for i, user := range sampleUsers {
		// Get balance from TigerBeetle
		balance, err := tbClient.GetBalance(user.TigerBeetleAccountID)
		if err != nil {
			log.Printf("  %d. ⚠️  %s - Failed to get balance: %v", i+1, user.Email, err)
			continue
		}

		// Convert cents to dollars
		balanceUSD := float64(balance) / 100.0

		// Get transaction count for this user
		var userTxCount int64
		db.Model(&models.Transaction{}).
			Where("user_id = ? OR recipient_user_id = ?", user.ID, user.ID).
			Count(&userTxCount)

		log.Printf("  %d. ✅ %s", i+1, user.Email)
		log.Printf("     Account: %s", user.AccountNumber)
		log.Printf("     Balance: $%.2f USD (%d cents)", balanceUSD, balance)
		log.Printf("     TigerBeetle ID: %d", user.TigerBeetleAccountID)
		log.Printf("     Transactions: %d", userTxCount)
	}

	// Check Isabel Hernández specifically
	log.Println("----------------------------------------------------------------")
	log.Println("🔍 Checking Test User (Isabel Hernández):")
	log.Println("----------------------------------------------------------------")

	var isabel models.User
	if err := db.Where("email = ?", "ihernandez@email.com").First(&isabel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Println("  ⚠️  Isabel Hernández not found in database")
		} else {
			return fmt.Errorf("failed to find Isabel: %w", err)
		}
	} else {
		balance, err := tbClient.GetBalance(isabel.TigerBeetleAccountID)
		if err != nil {
			log.Printf("  ⚠️  Failed to get Isabel's balance from TigerBeetle: %v", err)
		} else {
			balanceUSD := float64(balance) / 100.0
			log.Printf("  ✅ Email: %s", isabel.Email)
			log.Printf("     Account: %s", isabel.AccountNumber)
			log.Printf("     Balance: $%.2f USD (%d cents)", balanceUSD, balance)
			log.Printf("     TigerBeetle ID: %d", isabel.TigerBeetleAccountID)

			// Get her transactions
			var isabelTxCount int64
			db.Model(&models.Transaction{}).
				Where("user_id = ? OR recipient_user_id = ?", isabel.ID, isabel.ID).
				Count(&isabelTxCount)
			log.Printf("     Transactions: %d", isabelTxCount)

			// Show recent transactions
			if isabelTxCount > 0 {
				log.Println("     Recent Transactions:")
				var recentTx []models.Transaction
				db.Where("user_id = ? OR recipient_user_id = ?", isabel.ID, isabel.ID).
					Order("created_at DESC").
					Limit(3).
					Find(&recentTx)

				for _, tx := range recentTx {
					amountUSD := float64(tx.Amount) / 100.0
					log.Printf("       - %s: $%.2f - %s", tx.Type, amountUSD, tx.Description)
				}
			}
		}
	}

	log.Println("================================================================")
	log.Println("✅ DIAGNOSTICS COMPLETE")
	log.Println("================================================================")
	log.Println("")

	return nil
}
