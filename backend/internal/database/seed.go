package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/hlabs/banking-system/internal/models"
	"github.com/hlabs/banking-system/internal/tigerbeetle"
	"github.com/hlabs/banking-system/pkg/utils"
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

// TestDataFile represents the structure of the test data JSON
type TestDataFile struct {
	Users []TestUser `json:"users"`
}

// Seed populates the database with test users from datos-prueba-HNL.json
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

	// Parse JSON
	var testData TestDataFile
	if err := json.Unmarshal(fileData, &testData); err != nil {
		return fmt.Errorf("failed to parse test data JSON: %w", err)
	}

	totalUsers := len(testData.Users)
	log.Printf("📊 Found %d test users to seed", totalUsers)
	log.Println("----------------------------------------------------------------")
	log.Println("🚀 Creating users... (this may take a moment)")
	log.Println("----------------------------------------------------------------")

	// Create users
	successCount := 0
	failCount := 0
	startTime := time.Now()

	for i, testUser := range testData.Users {
		// Progress indicator - show every 10% or at specific milestones
		progress := float64(i+1) / float64(totalUsers) * 100
		showProgress := (i+1)%10 == 0 || i == 0 || i == totalUsers-1

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testUser.Password), bcrypt.DefaultCost)
		if err != nil {
			if showProgress {
				log.Printf("❌ [%d/%d] %.0f%% - Failed: %s (password hash error)", i+1, totalUsers, progress, testUser.Email)
			}
			failCount++
			continue
		}

		// Generate unique TigerBeetle account ID
		tbAccountID := utils.GenerateAccountID()

		// Create TigerBeetle account
		if err := tbClient.CreateAccount(tbAccountID); err != nil {
			if showProgress {
				log.Printf("❌ [%d/%d] %.0f%% - Failed: %s (TigerBeetle error)", i+1, totalUsers, progress, testUser.Email)
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
				log.Printf("❌ [%d/%d] %.0f%% - Failed: %s (database error)", i+1, totalUsers, progress, testUser.Email)
			}
			failCount++
			continue
		}

		successCount++

		// Show progress only at milestones
		if showProgress {
			log.Printf("✅ [%d/%d] %.0f%% - Created: %s (TB Account: %d)",
				i+1, totalUsers, progress, testUser.FullName, tbAccountID)
		}
	}

	duration := time.Since(startTime)

	// Summary
	log.Println("================================================================")
	log.Println("🌱 DATABASE SEEDING COMPLETED")
	log.Println("================================================================")
	log.Printf("   Total users processed: %d", totalUsers)
	log.Printf("   ✅ Successfully created: %d users", successCount)
	if failCount > 0 {
		log.Printf("   ❌ Failed: %d users", failCount)
	}
	log.Printf("   ⏱️  Time elapsed: %v", duration)
	log.Println("================================================================")

	if successCount == 0 && failCount > 0 {
		return fmt.Errorf("all user creations failed")
	}

	return nil
}
