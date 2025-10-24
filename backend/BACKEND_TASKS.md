# Backend Development Tasks - HLABS Banking System

## Overview
This document contains the detailed task breakdown for implementing the Go backend for the HLABS online banking system. Each task is designed to be actionable and testable.

**Tech Stack**: Go 1.21+, Gin, GORM, PostgreSQL, TigerBeetle, JWT, MCP

---

## Phase 1: Project Initialization (30 min)

### 1.1 Go Module Setup
- [ ] Navigate to `backend/` directory
- [ ] Initialize Go module: `go mod init github.com/[username]/hlabs-banking`
- [ ] Install core dependencies:
  ```bash
  go get -u github.com/gin-gonic/gin
  go get -u github.com/gin-contrib/cors
  go get -u gorm.io/gorm
  go get -u gorm.io/driver/postgres
  go get -u github.com/golang-jwt/jwt/v5
  go get -u golang.org/x/crypto/bcrypt
  go get -u github.com/joho/godotenv
  go get -u github.com/google/uuid
  ```

### 1.2 Environment Configuration
- [ ] Create `.env.example` with all required variables:
  - PostgreSQL connection string
  - TigerBeetle connection details
  - JWT secret key
  - Server port
  - OpenRouter API key (for MCP)
- [ ] Create `.env` file (gitignored) for local development

**Deliverable**: Go module initialized with dependencies

---

## Phase 2: Configuration & Database Connection (1 hour)

### 2.1 Configuration Module
- [ ] Create `internal/config/config.go`:
  - Struct for configuration values
  - Load from environment variables
  - Validation for required fields
  - Default values where appropriate

**Code Structure**:
```go
type Config struct {
    ServerPort        string
    PostgresDSN       string
    TigerBeetleHost   string
    TigerBeetlePort   string
    JWTSecret         string
    OpenRouterAPIKey  string
}

func Load() (*Config, error)
```

### 2.2 PostgreSQL Connection
- [ ] Create `internal/database/postgres.go`:
  - GORM database connection
  - Connection pooling configuration
  - Ping/health check
  - Auto-migration trigger

**Code Structure**:
```go
func Connect(dsn string) (*gorm.DB, error)
func Migrate(db *gorm.DB) error
```

### 2.3 User Model
- [ ] Create `internal/models/user.go`:
  - User struct with GORM tags
  - Fields: ID (UUID), Email, PasswordHash, FullName, TigerBeetleAccountID, CreatedAt, UpdatedAt
  - JSON tags for API responses (exclude password)
  - Validation tags

**Code Structure**:
```go
type User struct {
    ID                    uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Email                 string    `gorm:"uniqueIndex;not null"`
    PasswordHash          string    `gorm:"not null" json:"-"`
    FullName              string    `gorm:"not null"`
    TigerBeetleAccountID  uint64    `gorm:"not null"`
    CreatedAt             time.Time
    UpdatedAt             time.Time
}
```

**Deliverable**: Working PostgreSQL connection with User model

---

## Phase 3: Authentication System (2-3 hours)

### 3.1 JWT Service
- [ ] Create `internal/auth/jwt.go`:
  - GenerateToken(userID, email) - create JWT with claims
  - ValidateToken(tokenString) - parse and validate JWT
  - ExtractUserID(tokenString) - get user ID from token
  - Token expiration: 24 hours

**Code Structure**:
```go
type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

func GenerateToken(userID, email, secret string) (string, error)
func ValidateToken(tokenString, secret string) (*Claims, error)
```

### 3.2 Authentication Handlers
- [ ] Create `internal/auth/handlers.go`:
  - **Register Handler**:
    - Validate email format and password strength
    - Check if email already exists
    - Hash password with bcrypt
    - Create TigerBeetle account (placeholder for now)
    - Save user to PostgreSQL
    - Return JWT token
  - **Login Handler**:
    - Validate credentials
    - Compare password hash
    - Generate JWT token
    - Return user info + token
  - **Logout Handler** (optional):
    - Client-side token removal
    - Return success message

**Request/Response Structs**:
```go
type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    FullName string `json:"full_name" binding:"required"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
    Token string `json:"token"`
    User  UserDTO `json:"user"`
}

type UserDTO struct {
    ID       string `json:"id"`
    Email    string `json:"email"`
    FullName string `json:"full_name"`
}
```

### 3.3 Authentication Middleware
- [ ] Create `internal/middleware/auth.go`:
  - Extract JWT from Authorization header (Bearer token)
  - Validate token
  - Extract user ID and attach to Gin context
  - Return 401 if invalid/missing token

**Code Structure**:
```go
func AuthMiddleware(secret string) gin.HandlerFunc
```

### 3.4 Utility Functions
- [ ] Create `internal/utils/response.go`:
  - Standard error response format
  - Success response format
  - HTTP status code helpers

**Code Structure**:
```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
}

type SuccessResponse struct {
    Data    interface{} `json:"data"`
    Message string      `json:"message,omitempty"`
}

func ErrorJSON(c *gin.Context, status int, message string)
func SuccessJSON(c *gin.Context, status int, data interface{})
```

**Deliverable**: Working authentication API (register, login)

---

## Phase 4: TigerBeetle Integration (3-4 hours)

### 4.1 TigerBeetle Client Wrapper
- [ ] Install TigerBeetle Go client:
  ```bash
  go get github.com/tigerbeetle/tigerbeetle-go
  ```
- [ ] Create `internal/tigerbeetle/client.go`:
  - Client initialization
  - Connection management
  - Health check
  - Error handling wrapper

**Code Structure**:
```go
type Client struct {
    client *tigerbeetle.Client
}

func NewClient(address string) (*Client, error)
func (c *Client) Close() error
func (c *Client) HealthCheck() error
```

### 4.2 Account Service
- [ ] Create `internal/account/service.go`:
  - **CreateAccount(userID)**:
    - Generate unique TigerBeetle account ID
    - Create account with ledger ID 1 (user accounts)
    - Set initial balance to 0
    - Return account ID
  - **GetBalance(accountID)**:
    - Query TigerBeetle for account
    - Return debits, credits, and net balance
  - **GetAccountInfo(accountID)**:
    - Return full account details

**TigerBeetle Account Structure**:
```go
// Account codes:
// - User accounts: ledger 1
// - System bank account: ledger 2
// Account flags: DebitsMustNotExceedCredits for user accounts
```

### 4.3 Account Handlers
- [ ] Create `internal/account/handlers.go`:
  - **GET /api/accounts/me**:
    - Get authenticated user from context
    - Fetch TigerBeetle account ID from user record
    - Return account info
  - **GET /api/accounts/balance**:
    - Get authenticated user
    - Query balance from TigerBeetle
    - Return balance

**Response Struct**:
```go
type AccountResponse struct {
    AccountID string  `json:"account_id"`
    Balance   float64 `json:"balance"`
    Currency  string  `json:"currency"`
}
```

### 4.4 Update Registration Flow
- [ ] Modify `internal/auth/handlers.go`:
  - Call account service to create TigerBeetle account during registration
  - Store returned account ID in PostgreSQL user record
  - Handle TigerBeetle errors gracefully

**Deliverable**: Account creation and balance retrieval working

---

## Phase 5: Transaction Operations (4-5 hours)

### 5.1 Study TigerBeetle Transfer Model
- [ ] Research TigerBeetle transfers documentation
- [ ] Understand flags: `linked`, `pending`, `post_pending_transfer`
- [ ] Understand two-phase commits for confirmations

### 5.2 Transaction Service
- [ ] Create `internal/transaction/service.go`:

  **Deposit(userAccountID, amount)**:
  - Create transfer from system account → user account
  - Amount in cents (multiply by 100)
  - Ledger ID: 1
  - Return transfer ID

  **Withdraw(userAccountID, amount)**:
  - Validate balance >= amount
  - Create transfer from user account → system account
  - Return transfer ID or insufficient funds error

  **Transfer(fromAccountID, toAccountID, amount)**:
  - Validate both accounts exist
  - Validate balance >= amount
  - Create transfer between user accounts
  - Return transfer ID

  **GetHistory(accountID, page, limit)**:
  - Query transfers involving account
  - Support pagination
  - Return both incoming and outgoing transfers
  - Calculate running balance

**Code Structures**:
```go
type TransactionService struct {
    tbClient *tigerbeetle.Client
    db       *gorm.DB
}

type Transfer struct {
    ID              string
    FromAccountID   uint64
    ToAccountID     uint64
    Amount          uint64  // in cents
    Timestamp       time.Time
    Type            string  // "deposit", "withdraw", "transfer"
}

func (s *TransactionService) Deposit(ctx context.Context, accountID uint64, amount float64) error
func (s *TransactionService) Withdraw(ctx context.Context, accountID uint64, amount float64) error
func (s *TransactionService) Transfer(ctx context.Context, fromID, toID uint64, amount float64) error
func (s *TransactionService) GetHistory(accountID uint64, page, limit int) ([]Transfer, error)
```

### 5.3 Transaction Handlers
- [ ] Create `internal/transaction/handlers.go`:

  **POST /api/transactions/deposit**:
  - Extract amount from request
  - Validate amount > 0
  - Call service.Deposit()
  - Return success + new balance

  **POST /api/transactions/withdraw**:
  - Extract amount from request
  - Validate amount > 0
  - Get user's TigerBeetle account
  - Call service.Withdraw()
  - Handle insufficient funds error
  - Return success + new balance

  **POST /api/transactions/transfer**:
  - Extract destination account ID and amount
  - Validate destination exists
  - Validate user owns source account
  - Call service.Transfer()
  - Return success + new balance

  **GET /api/transactions/history**:
  - Extract page and limit from query params
  - Get user's account
  - Call service.GetHistory()
  - Return paginated results

**Request Structs**:
```go
type DepositRequest struct {
    Amount float64 `json:"amount" binding:"required,gt=0"`
}

type WithdrawRequest struct {
    Amount float64 `json:"amount" binding:"required,gt=0"`
}

type TransferRequest struct {
    ToAccountID uint64  `json:"to_account_id" binding:"required"`
    Amount      float64 `json:"amount" binding:"required,gt=0"`
}

type HistoryResponse struct {
    Transactions []TransactionDTO `json:"transactions"`
    Page         int              `json:"page"`
    Limit        int              `json:"limit"`
    Total        int              `json:"total"`
}
```

### 5.4 Validation & Error Handling
- [ ] Validate account ownership before operations
- [ ] Handle TigerBeetle errors (account not found, insufficient funds)
- [ ] Add amount limits (min: $0.01, max: $10,000 per transaction)
- [ ] Return meaningful error messages

**Deliverable**: Complete transaction API with TigerBeetle

---

## Phase 6: AI Chat Integration (3-4 hours)

### 6.1 MCP Client Setup
- [ ] Install MCP Go SDK:
  ```bash
  go get github.com/modelcontextprotocol/go-sdk
  ```
- [ ] Research OpenRouter API integration
- [ ] Create `internal/chat/client.go`:
  - MCP client wrapper
  - OpenRouter API calls
  - Message formatting
  - Streaming support (optional)

### 6.2 Intent Recognition
- [ ] Create `internal/chat/intents.go`:
  - Define intent types (enum/constants):
    - CHECK_BALANCE
    - DEPOSIT
    - WITHDRAW
    - TRANSFER
    - VIEW_HISTORY
    - UNKNOWN
  - Intent extraction from user message
  - Parameter extraction (amount, account ID)

**Code Structure**:
```go
type Intent struct {
    Type       string
    Amount     float64
    ToAccount  uint64
    Confidence float64
}

func ParseIntent(message string, aiResponse string) (*Intent, error)
```

### 6.3 Chat Service
- [ ] Create `internal/chat/service.go`:
  - **ProcessMessage(userMessage, userID)**:
    - Send message to MCP/OpenRouter
    - Parse AI response for intent
    - Extract parameters
    - Return intent + confirmation prompt

  - **ExecuteIntent(intent, userID)**:
    - Call appropriate service method
    - Handle errors
    - Format response message

  - **FormatResponse(result)**:
    - Create user-friendly message
    - Include balance, transaction ID, etc.

**Code Structure**:
```go
type ChatService struct {
    mcpClient     *mcp.Client
    accountSvc    *account.Service
    transactionSvc *transaction.Service
}

func (s *ChatService) ProcessMessage(ctx context.Context, message, userID string) (*ChatResponse, error)
func (s *ChatService) ExecuteIntent(ctx context.Context, intent *Intent, userID string) (*ExecutionResult, error)
```

### 6.4 Chat Handler
- [ ] Create `internal/chat/handlers.go`:
  - **POST /api/chat**:
    - Extract message from request
    - Get authenticated user
    - Process message through ChatService
    - If critical operation (transfer, withdraw):
      - Return confirmation prompt
      - Wait for user confirmation
    - Execute intent if confirmed
    - Return formatted response

**Request/Response Structs**:
```go
type ChatRequest struct {
    Message    string `json:"message" binding:"required"`
    Confirmed  bool   `json:"confirmed"`
    IntentData string `json:"intent_data,omitempty"` // Serialized intent for confirmation
}

type ChatResponse struct {
    Message         string      `json:"message"`
    NeedsConfirmation bool      `json:"needs_confirmation"`
    Intent          *Intent     `json:"intent,omitempty"`
    Result          interface{} `json:"result,omitempty"`
}
```

### 6.5 Safety & Validation
- [ ] Require confirmation for:
  - Withdrawals
  - Transfers
- [ ] Validate extracted amounts and account IDs
- [ ] Handle ambiguous requests gracefully
- [ ] Implement rate limiting (optional)

**Deliverable**: Working AI chat interface with MCP

---

## Phase 7: Main Server & Routes (1 hour)

### 7.1 Main Server File
- [ ] Create `cmd/server/main.go`:
  - Load configuration
  - Initialize database connections (PostgreSQL, TigerBeetle)
  - Initialize services
  - Setup Gin router with middleware
  - Define all routes
  - Start server

**Code Structure**:
```go
func main() {
    // Load config
    cfg := config.Load()

    // Connect to PostgreSQL
    db := database.Connect(cfg.PostgresDSN)
    database.Migrate(db)

    // Connect to TigerBeetle
    tbClient := tigerbeetle.NewClient(cfg.TigerBeetleHost)

    // Initialize services
    authSvc := auth.NewService(db, tbClient)
    accountSvc := account.NewService(tbClient, db)
    txnSvc := transaction.NewService(tbClient, db)
    chatSvc := chat.NewService(mcpClient, accountSvc, txnSvc)

    // Setup router
    r := gin.Default()
    r.Use(cors.Default())
    r.Use(gin.Logger())
    r.Use(gin.Recovery())

    // Routes
    setupRoutes(r, authSvc, accountSvc, txnSvc, chatSvc, cfg.JWTSecret)

    // Start server
    r.Run(":" + cfg.ServerPort)
}
```

### 7.2 Route Definition
- [ ] Define route groups:
  ```go
  // Public routes
  auth := r.Group("/api/auth")
  {
      auth.POST("/register", authHandlers.Register)
      auth.POST("/login", authHandlers.Login)
      auth.POST("/logout", authHandlers.Logout)
  }

  // Protected routes
  api := r.Group("/api")
  api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
  {
      // Accounts
      api.GET("/accounts/me", accountHandlers.GetAccount)
      api.GET("/accounts/balance", accountHandlers.GetBalance)

      // Transactions
      api.POST("/transactions/deposit", txnHandlers.Deposit)
      api.POST("/transactions/withdraw", txnHandlers.Withdraw)
      api.POST("/transactions/transfer", txnHandlers.Transfer)
      api.GET("/transactions/history", txnHandlers.GetHistory)

      // Chat
      api.POST("/chat", chatHandlers.ProcessMessage)
  }
  ```

### 7.3 CORS Configuration
- [ ] Configure CORS middleware:
  - Allow frontend origin (http://localhost:5173 for dev)
  - Allow credentials
  - Allow required headers

**Deliverable**: Complete server with all routes

---

## Phase 8: Docker & Deployment (2 hours)

### 8.1 Dockerfile
- [ ] Create `backend/Dockerfile`:
  - **Stage 1: Builder**
    - Use golang:1.21-alpine
    - Copy go.mod and go.sum
    - Download dependencies
    - Copy source code
    - Build binary
  - **Stage 2: Production**
    - Use alpine:latest
    - Copy binary from builder
    - Expose port 8080
    - Run binary

**Dockerfile Example**:
```dockerfile
# Builder stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Production stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

### 8.2 Docker Compose Integration
- [ ] Update root `docker-compose.yml`:
  - Add backend service
  - Link to PostgreSQL and TigerBeetle
  - Configure environment variables
  - Add health checks
  - Configure restart policy

**Example**:
```yaml
services:
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - POSTGRES_DSN=postgresql://user:password@postgres:5432/banking
      - TIGERBEETLE_HOST=tigerbeetle
      - TIGERBEETLE_PORT=3000
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - postgres
      - tigerbeetle
    restart: unless-stopped
```

### 8.3 Data Seeding Script
- [ ] Create `backend/scripts/seed.go`:
  - Read `datos-prueba-HNL.json`
  - Create users in PostgreSQL
  - Create TigerBeetle accounts for each user
  - Hash passwords
  - Run as standalone command

**Usage**:
```bash
go run scripts/seed.go
```

**Deliverable**: Dockerized backend service

---

## Phase 9: Testing & Validation (2 hours)

### 9.1 Manual Endpoint Testing
- [ ] Test with curl/Postman:
  - POST /api/auth/register
  - POST /api/auth/login (verify JWT)
  - GET /api/accounts/balance (with JWT)
  - POST /api/transactions/deposit
  - POST /api/transactions/withdraw (test insufficient funds)
  - POST /api/transactions/transfer
  - GET /api/transactions/history
  - POST /api/chat (test balance check)

### 9.2 Integration Testing
- [ ] Test full flow:
  1. Register user
  2. Login
  3. Check balance (should be 0)
  4. Deposit $100
  5. Check balance (should be $100)
  6. Withdraw $30
  7. Transfer $20 to another user
  8. Check balance (should be $50)
  9. View history (should show 3 transactions)

### 9.3 Error Scenarios
- [ ] Test error handling:
  - Invalid credentials
  - Duplicate email registration
  - Insufficient funds withdrawal
  - Transfer to non-existent account
  - Invalid JWT token
  - Missing required fields

### 9.4 Docker Testing
- [ ] Test with `docker-compose up`:
  - All services start successfully
  - Backend connects to both databases
  - Endpoints accessible from host machine
  - Data persists after container restart

**Deliverable**: Tested and validated backend

---

## Phase 10: Documentation (1 hour)

### 10.1 API Documentation
- [ ] Create `backend/API.md`:
  - List all endpoints
  - Request/response examples
  - Authentication requirements
  - Error codes

### 10.2 Backend README
- [ ] Create `backend/README.md`:
  - Setup instructions
  - Environment variables
  - Running locally
  - Running with Docker
  - Architecture overview

### 10.3 Code Comments
- [ ] Add comments to:
  - Complex business logic
  - TigerBeetle operations
  - Security-critical code
  - Public functions/methods

**Deliverable**: Well-documented backend

---

## Completion Checklist

### Must Have (P0)
- [ ] Authentication working (register, login)
- [ ] PostgreSQL connection and user storage
- [ ] TigerBeetle integration (accounts, transfers)
- [ ] All transaction operations (deposit, withdraw, transfer)
- [ ] Transaction history
- [ ] AI chat integration with MCP
- [ ] All endpoints tested
- [ ] Docker setup complete

### Should Have (P1)
- [ ] Input validation on all endpoints
- [ ] Proper error handling
- [ ] CORS configuration
- [ ] Request logging
- [ ] Environment configuration
- [ ] Data seeding script

### Nice to Have (P2)
- [ ] Unit tests
- [ ] Rate limiting
- [ ] Swagger/OpenAPI documentation
- [ ] Structured logging
- [ ] Metrics/monitoring hooks

---

## Estimated Timeline

| Phase | Tasks | Duration |
|-------|-------|----------|
| 1. Initialization | Go setup, dependencies | 30 min |
| 2. Config & DB | Config, PostgreSQL, models | 1 hour |
| 3. Authentication | JWT, register, login, middleware | 2-3 hours |
| 4. TigerBeetle | Client, accounts, integration | 3-4 hours |
| 5. Transactions | Deposit, withdraw, transfer, history | 4-5 hours |
| 6. AI Chat | MCP, intent parsing, chat API | 3-4 hours |
| 7. Server Setup | Main server, routes | 1 hour |
| 8. Docker | Dockerfile, compose, seeding | 2 hours |
| 9. Testing | Manual, integration, error cases | 2 hours |
| 10. Documentation | API docs, README, comments | 1 hour |

**Total**: ~18-24 hours

---

## Risk Mitigation

### TigerBeetle Complexity
- **Risk**: Unfamiliar with TigerBeetle's accounting model
- **Mitigation**:
  - Read docs thoroughly before coding
  - Start with simple account creation
  - Test transfers extensively
  - Keep system account logic simple

### MCP Integration
- **Risk**: MCP SDK may have compatibility issues
- **Mitigation**:
  - Test MCP connection early
  - Have fallback: direct OpenAI API
  - Keep intent parsing simple initially

### Time Constraints
- **Risk**: May run out of time
- **Mitigation**:
  - Focus on P0 features first
  - Skip P2 features if needed
  - Simplify AI chat if necessary

---

## Next Steps

**START HERE**:
1. ✅ Create directory structure
2. ✅ Create this task file
3. ⏭️ Initialize Go module
4. ⏭️ Create .env.example
5. ⏭️ Implement config loading

**Remember**: Commit after each completed phase!
