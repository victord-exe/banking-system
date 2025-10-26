package account

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/hlabs/banking-system/internal/models"
	"github.com/hlabs/banking-system/internal/tigerbeetle"
	"gorm.io/gorm"
)

// Service handles account-related business logic
type Service struct {
	db       *gorm.DB
	tbClient *tigerbeetle.Client
}

// NewService creates a new account service
func NewService(db *gorm.DB, tbClient *tigerbeetle.Client) *Service {
	return &Service{
		db:       db,
		tbClient: tbClient,
	}
}

// GetUserByID retrieves a user by their ID
func (s *Service) GetUserByID(userID string) (*models.User, error) {
	var user models.User

	// Parse UUID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	// Query database
	if err := s.db.First(&user, "id = ?", uid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}

// GetBalance retrieves the balance for a user's TigerBeetle account
func (s *Service) GetBalance(userID string) (int64, error) {
	// Get user to retrieve their TigerBeetle account ID
	user, err := s.GetUserByID(userID)
	if err != nil {
		return 0, err
	}

	// Query TigerBeetle for balance
	balance, err := s.tbClient.GetBalance(user.TigerBeetleAccountID)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance from TigerBeetle: %w", err)
	}

	return balance, nil
}
