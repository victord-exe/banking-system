package models

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	tb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeWithdraw TransactionType = "withdraw"
	TransactionTypeTransfer TransactionType = "transfer"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
)

// Transaction represents a financial transaction history record
// This is the PostgreSQL audit log that tracks all TigerBeetle operations
type Transaction struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

	// User references (PostgreSQL users)
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index:idx_transactions_user_created,priority:1" json:"user_id"`
	User            *User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	RecipientUserID *uuid.UUID `gorm:"type:uuid;index:idx_transactions_recipient_user_id" json:"recipient_user_id,omitempty"`
	RecipientUser   *User      `gorm:"foreignKey:RecipientUserID;constraint:OnDelete:SET NULL" json:"recipient_user,omitempty"`

	// Transaction details
	Type   TransactionType   `gorm:"type:varchar(10);not null;check:type IN ('deposit','withdraw','transfer');index:idx_transactions_type" json:"type"`
	Amount int64             `gorm:"not null;check:amount > 0" json:"amount"` // Amount in cents
	Status TransactionStatus `gorm:"type:varchar(10);not null;default:'pending';check:status IN ('pending','completed','failed');index:idx_transactions_status" json:"status"`

	// TigerBeetle references (stored as BIGINT, no FK - different database)
	DebitAccountID  uint64 `gorm:"not null" json:"debit_account_id"`
	CreditAccountID uint64 `gorm:"not null" json:"credit_account_id"`

	// TigerBeetle transfer tracking (uint128 stored as hex string)
	TigerBeetleTransferID string `gorm:"type:varchar(32);not null;uniqueIndex" json:"tigerbeetle_transfer_id"`

	// Optional fields
	Description string         `gorm:"type:text" json:"description,omitempty"`
	Metadata    datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata,omitempty"`

	// Timestamps
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_transactions_user_created,priority:2" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete support
}

// TableName specifies the table name for the Transaction model
func (Transaction) TableName() string {
	return "transactions"
}

// BeforeCreate hook to set default values
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.Status == "" {
		t.Status = TransactionStatusPending
	}
	return nil
}

// SetTigerBeetleTransferID converts TigerBeetle Uint128 to hex string and stores it
func (t *Transaction) SetTigerBeetleTransferID(id tb_types.Uint128) {
	// Convert Uint128 to bytes and then to hex string
	bytes := make([]byte, 16)
	bigInt := id.BigInt()
	bigInt.FillBytes(bytes)
	t.TigerBeetleTransferID = hex.EncodeToString(bytes)
}

// GetTigerBeetleTransferID converts hex string back to TigerBeetle Uint128
func (t *Transaction) GetTigerBeetleTransferID() (tb_types.Uint128, error) {
	bytes, err := hex.DecodeString(t.TigerBeetleTransferID)
	if err != nil {
		return tb_types.Uint128{}, err
	}

	// Convert slice to fixed-size array (BytesToUint128 requires [16]byte, not []byte)
	if len(bytes) != 16 {
		return tb_types.Uint128{}, fmt.Errorf("invalid transfer ID length: expected 16 bytes, got %d", len(bytes))
	}

	var fixedBytes [16]byte
	copy(fixedBytes[:], bytes)

	return tb_types.BytesToUint128(fixedBytes), nil
}

// TransactionDTO is the data transfer object for transaction information
// Used for API responses with enriched data
type TransactionDTO struct {
	ID                    uuid.UUID         `json:"id"`
	UserID                uuid.UUID         `json:"user_id"`
	RecipientUserID       *uuid.UUID        `json:"recipient_user_id,omitempty"`
	Type                  TransactionType   `json:"type"`
	Amount                int64             `json:"amount"` // Amount in cents
	AmountFormatted       string            `json:"amount_formatted"`
	Status                TransactionStatus `json:"status"`
	DebitAccountID        uint64            `json:"debit_account_id"`
	CreditAccountID       uint64            `json:"credit_account_id"`
	TigerBeetleTransferID string            `json:"tigerbeetle_transfer_id"`
	Description           string            `json:"description,omitempty"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`

	// Enriched fields (from JOINs)
	RecipientEmail string `json:"recipient_email,omitempty"`
	RecipientName  string `json:"recipient_name,omitempty"`
}

// ToDTO converts a Transaction to TransactionDTO with enriched information
func (t *Transaction) ToDTO() TransactionDTO {
	dto := TransactionDTO{
		ID:                    t.ID,
		UserID:                t.UserID,
		RecipientUserID:       t.RecipientUserID,
		Type:                  t.Type,
		Amount:                t.Amount,
		AmountFormatted:       formatAmount(t.Amount),
		Status:                t.Status,
		DebitAccountID:        t.DebitAccountID,
		CreditAccountID:       t.CreditAccountID,
		TigerBeetleTransferID: t.TigerBeetleTransferID,
		Description:           t.Description,
		CreatedAt:             t.CreatedAt,
		UpdatedAt:             t.UpdatedAt,
	}

	// Enrich with recipient information if available
	if t.RecipientUser != nil {
		dto.RecipientEmail = t.RecipientUser.Email
		dto.RecipientName = t.RecipientUser.FullName
	}

	return dto
}

// formatAmount converts cents to dollar string (e.g., 12345 -> "$123.45")
func formatAmount(cents int64) string {
	dollars := float64(cents) / 100.0
	return fmt.Sprintf("$%.2f", dollars)
}

// Uint128ToHex converts TigerBeetle Uint128 to hex string for storage
func Uint128ToHex(u tb_types.Uint128) string {
	bytes := make([]byte, 16)
	bigInt := u.BigInt()
	bigInt.FillBytes(bytes)
	return hex.EncodeToString(bytes)
}

// HexToUint128 converts hex string back to TigerBeetle Uint128
func HexToUint128(s string) (tb_types.Uint128, error) {
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return tb_types.Uint128{}, err
	}

	// Convert slice to fixed-size array (BytesToUint128 requires [16]byte, not []byte)
	if len(bytes) != 16 {
		return tb_types.Uint128{}, fmt.Errorf("invalid hex length: expected 32 characters (16 bytes), got %d bytes", len(bytes))
	}

	var fixedBytes [16]byte
	copy(fixedBytes[:], bytes)

	return tb_types.BytesToUint128(fixedBytes), nil
}
