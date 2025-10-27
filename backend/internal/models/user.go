package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the banking system
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"` // Never expose password in JSON
	FullName  string    `gorm:"not null" json:"full_name"`

	// TigerBeetle account ID - links to financial account
	TigerBeetleAccountID uint64 `gorm:"not null;uniqueIndex" json:"tigerbeetle_account_id"`

	// Account number from test data (e.g., "4001-6588-5247-0001")
	// This is used to link transactions from the JSON test data
	AccountNumber string `gorm:"index" json:"account_number,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// UserDTO is the data transfer object for user information (safe for API responses)
type UserDTO struct {
	ID                   uuid.UUID `json:"id"`
	Email                string    `json:"email"`
	FullName             string    `json:"full_name"`
	TigerBeetleAccountID uint64    `json:"tigerbeetle_account_id"`
	AccountNumber        string    `json:"account_number,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}

// ToDTO converts a User to UserDTO
func (u *User) ToDTO() UserDTO {
	return UserDTO{
		ID:                   u.ID,
		Email:                u.Email,
		FullName:             u.FullName,
		TigerBeetleAccountID: u.TigerBeetleAccountID,
		AccountNumber:        u.AccountNumber,
		CreatedAt:            u.CreatedAt,
	}
}
