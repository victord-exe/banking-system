package transaction

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/hlabs/banking-system/internal/models"
	"gorm.io/gorm"
)

// Repository handles database operations for transactions
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new transaction repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create saves a new transaction record to the database
func (r *Repository) Create(tx *models.Transaction) error {
	if err := r.db.Create(tx).Error; err != nil {
		return fmt.Errorf("failed to create transaction record: %w", err)
	}
	return nil
}

// GetByID retrieves a single transaction by its ID
func (r *Repository) GetByID(id uuid.UUID) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.
		Preload("User").
		Preload("RecipientUser").
		First(&tx, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("failed to retrieve transaction: %w", err)
	}

	return &tx, nil
}

// GetByUserID retrieves paginated transactions for a specific user
// This query fetches all transactions where the user is the initiator
func (r *Repository) GetByUserID(userID uuid.UUID, page, limit int) ([]models.Transaction, error) {
	var transactions []models.Transaction

	// Calculate offset for pagination
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	// Query with JOIN to load recipient user information
	err := r.db.
		Preload("User").
		Preload("RecipientUser").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transactions: %w", err)
	}

	return transactions, nil
}

// GetAllByUserID retrieves ALL transactions involving a user (sent or received)
// This includes transactions where the user is either the sender OR recipient
func (r *Repository) GetAllByUserID(userID uuid.UUID, page, limit int) ([]models.Transaction, error) {
	var transactions []models.Transaction

	// Calculate offset for pagination
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	log.Printf("🔍 [Repository] GetAllByUserID called with userID=%s, page=%d, limit=%d, offset=%d", userID, page, limit, offset)
	log.Printf("🔍 [Repository] Query: WHERE user_id = '%s' OR recipient_user_id = '%s'", userID, userID)

	// Query for transactions where user is either sender or recipient
	err := r.db.
		Preload("User").
		Preload("RecipientUser").
		Where("user_id = ? OR recipient_user_id = ?", userID, userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	if err != nil {
		log.Printf("❌ [Repository] Query failed: %v", err)
		return nil, fmt.Errorf("failed to retrieve transactions: %w", err)
	}

	log.Printf("✅ [Repository] Query returned %d transactions", len(transactions))
	return transactions, nil
}

// GetByUserIDAndType retrieves paginated transactions filtered by type
func (r *Repository) GetByUserIDAndType(userID uuid.UUID, txType models.TransactionType, page, limit int) ([]models.Transaction, error) {
	var transactions []models.Transaction

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	err := r.db.
		Preload("User").
		Preload("RecipientUser").
		Where("user_id = ? AND type = ?", userID, txType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transactions by type: %w", err)
	}

	return transactions, nil
}

// CountByUserID returns the total number of transactions for a user
// Used for pagination metadata
func (r *Repository) CountByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Transaction{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	return count, nil
}

// CountAllByUserID returns the total number of transactions involving a user (sent or received)
func (r *Repository) CountAllByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Transaction{}).
		Where("user_id = ? OR recipient_user_id = ?", userID, userID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count all transactions: %w", err)
	}

	return count, nil
}

// UpdateStatus updates the status of a transaction
// Used when TigerBeetle operations complete or fail
func (r *Repository) UpdateStatus(id uuid.UUID, status models.TransactionStatus) error {
	err := r.db.Model(&models.Transaction{}).
		Where("id = ?", id).
		Update("status", status).Error

	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

// GetRecent retrieves the N most recent transactions for a user
// Useful for dashboard "recent activity" widgets
func (r *Repository) GetRecent(userID uuid.UUID, limit int) ([]models.Transaction, error) {
	var transactions []models.Transaction

	err := r.db.
		Preload("User").
		Preload("RecipientUser").
		Where("user_id = ? OR recipient_user_id = ?", userID, userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&transactions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve recent transactions: %w", err)
	}

	return transactions, nil
}

// GetByTigerBeetleTransferID retrieves a transaction by its TigerBeetle transfer ID
// Useful for reconciliation and debugging
func (r *Repository) GetByTigerBeetleTransferID(transferID string) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.
		Preload("User").
		Preload("RecipientUser").
		Where("tigerbeetle_transfer_id = ?", transferID).
		First(&tx).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("failed to retrieve transaction: %w", err)
	}

	return &tx, nil
}
