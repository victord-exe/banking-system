package account

import (
	"fmt"
	"log"

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
	log.Printf("🟡 [AccountService] GetBalance called for userID: %s", userID)

	// Get user to retrieve their TigerBeetle account ID
	log.Printf("🟡 [AccountService] Fetching user from database...")
	user, err := s.GetUserByID(userID)
	if err != nil {
		log.Printf("❌ [AccountService] Failed to get user: %v", err)
		return 0, err
	}
	log.Printf("🟡 [AccountService] User found: Email=%s, TigerBeetleAccountID=%d", user.Email, user.TigerBeetleAccountID)

	// Query TigerBeetle for balance
	log.Printf("🟡 [AccountService] Querying TigerBeetle for balance (AccountID: %d)...", user.TigerBeetleAccountID)
	balance, err := s.tbClient.GetBalance(user.TigerBeetleAccountID)
	if err != nil {
		log.Printf("❌ [AccountService] Failed to get balance from TigerBeetle: %v", err)
		return 0, fmt.Errorf("failed to get balance from TigerBeetle: %w", err)
	}
	log.Printf("✅ [AccountService] TigerBeetle returned balance: %d cents for account %d", balance, user.TigerBeetleAccountID)

	return balance, nil
}
