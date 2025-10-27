package database

import (
	"fmt"
	"log"
	"time"

	"github.com/hlabs/banking-system/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes a connection to PostgreSQL database
func Connect(dsn string) (*gorm.DB, error) {
	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)

	// Connect to database with UTF-8 support
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		// Ensure proper UTF-8 handling
		PrepareStmt: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Verify UTF-8 encoding is set correctly
	var clientEncoding string
	if err := db.Raw("SHOW client_encoding").Scan(&clientEncoding).Error; err != nil {
		log.Printf("⚠️  Warning: Could not verify client encoding: %v", err)
	} else {
		log.Printf("✅ PostgreSQL connection established (client_encoding=%s)", clientEncoding)
		if clientEncoding != "UTF8" {
			log.Printf("⚠️  WARNING: Client encoding is %s, expected UTF8 - special characters may not display correctly", clientEncoding)
		}
	}

	return db, nil
}

// Migrate runs database migrations
func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.User{},
		&models.Transaction{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("✅ Database migrations completed")

	return nil
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
