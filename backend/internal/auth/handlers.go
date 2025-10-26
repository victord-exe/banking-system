package auth

import (
	"log"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/hlabs/banking-system/internal/models"
	"github.com/hlabs/banking-system/internal/tigerbeetle"
	"github.com/hlabs/banking-system/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Handler handles authentication requests
type Handler struct {
	db        *gorm.DB
	tbClient  *tigerbeetle.Client
	jwtSecret string
}

// NewHandler creates a new auth handler
func NewHandler(db *gorm.DB, tbClient *tigerbeetle.Client, jwtSecret string) *Handler {
	return &Handler{
		db:        db,
		tbClient:  tbClient,
		jwtSecret: jwtSecret,
	}
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string         `json:"token"`
	User  models.UserDTO `json:"user"`
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate email format
	if !isValidEmail(req.Email) {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid email format")
		return
	}

	// Validate password strength
	if len(req.Password) < 6 {
		utils.RespondWithError(c, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		utils.RespondWithError(c, http.StatusConflict, "User with this email already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Generate unique TigerBeetle account ID
	tbAccountID := utils.GenerateAccountID()

	// Create TigerBeetle account
	if err := h.tbClient.CreateAccount(tbAccountID); err != nil {
		log.Printf("Error creating TigerBeetle account: %v", err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create bank account")
		return
	}

	// Create user in PostgreSQL
	user := models.User{
		Email:                req.Email,
		Password:             string(hashedPassword),
		FullName:             req.FullName,
		TigerBeetleAccountID: tbAccountID,
	}

	if err := h.db.Create(&user).Error; err != nil {
		log.Printf("Error creating user: %v", err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate JWT token
	token, err := GenerateToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to generate authentication token")
		return
	}

	// Return response
	response := AuthResponse{
		Token: token,
		User:  user.ToDTO(),
	}

	log.Printf("✅ User registered: %s (TB Account: %d)", user.Email, tbAccountID)

	utils.RespondWithSuccess(c, http.StatusCreated, response, "User registered successfully")
}

// Login handles user login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Find user by email
	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.RespondWithError(c, http.StatusUnauthorized, "Invalid email or password")
		} else {
			log.Printf("Error finding user: %v", err)
			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to authenticate")
		}
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Generate JWT token
	token, err := GenerateToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to generate authentication token")
		return
	}

	// Return response
	response := AuthResponse{
		Token: token,
		User:  user.ToDTO(),
	}

	log.Printf("✅ User logged in: %s", user.Email)

	utils.RespondWithSuccess(c, http.StatusOK, response, "Login successful")
}

// Logout handles user logout (client-side token removal)
func (h *Handler) Logout(c *gin.Context) {
	utils.RespondWithSuccess(c, http.StatusOK, nil, "Logout successful")
}

// Helper functions

// isValidEmail validates that the provided email string matches a valid email format.
// It uses a regex pattern to check for standard email structure (local@domain.tld).
// Returns true if the email is valid, false otherwise.
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
