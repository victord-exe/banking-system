# Work Plan - Online Banking System

## Overview
This document outlines the development roadmap for building the online banking system with Go backend, React frontend, TigerBeetle for financial operations, and AI chat integration via MCP.

**Timeline**: Maximum 2 days
**Start Date**: 2025-10-24

---

## Phase 1: Project Setup & Repository (0.5 hours)

### 1.1 GitHub Repository Setup ✓ FIRST TASK
- [ ] Create public GitHub repository: `banking-system` or similar
- [ ] Initialize with README.md (basic project description)
- [ ] Add .gitignore for Go and Node.js
- [ ] Create initial branch structure (main)
- [ ] Clone repository locally

### 1.2 Project Structure
- [ ] Create root directory structure:
  ```
  banking-system/
  ├── backend/
  ├── frontend/
  ├── docker-compose.yml
  ├── .env.example
  ├── README.md
  └── datos-prueba-HNL.json (copy from current location)
  ```

### 1.3 Documentation Setup
- [ ] Create README.md with project description
- [ ] Create .env.example with all required environment variables
- [ ] Copy CLAUDE.md and WORKPLAN.md to repository

**Deliverables**: GitHub repo URL, basic project structure

---

## Phase 2: Database Infrastructure (2-3 hours)

### 2.1 Docker Compose Foundation
- [ ] Create `docker-compose.yml` with:
  - PostgreSQL service (port 5432)
  - TigerBeetle service (port 3000)
  - Network configuration
  - Volume mounts for data persistence

### 2.2 PostgreSQL Setup
- [ ] Create `backend/migrations/` directory
- [ ] Design users table schema:
  ```sql
  CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    tigerbeetle_account_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  ```
- [ ] Create init.sql script for database initialization
- [ ] Test PostgreSQL connection via Docker

### 2.3 TigerBeetle Setup
- [ ] Study TigerBeetle documentation (accounts and transfers)
- [ ] Create TigerBeetle initialization script
- [ ] Define account structure for:
  - User accounts (ledger 1)
  - System accounts for deposits/withdrawals (ledger 2)
- [ ] Test TigerBeetle connection via Docker
- [ ] Create system accounts (bank account for deposits/withdrawals)

### 2.4 Test Database Setup
- [ ] Run `docker-compose up` for databases only
- [ ] Verify both databases are accessible
- [ ] Document connection strings in .env.example

**Deliverables**: Working docker-compose with PostgreSQL and TigerBeetle

---

## Phase 3: Backend Core - Authentication (3-4 hours)

### 3.1 Go Project Initialization
- [ ] Navigate to `backend/` directory
- [ ] Initialize Go module: `go mod init github.com/yourusername/banking-system`
- [ ] Install core dependencies:
  ```bash
  go get -u github.com/gin-gonic/gin
  go get -u gorm.io/gorm
  go get -u gorm.io/driver/postgres
  go get -u github.com/golang-jwt/jwt/v5
  go get -u golang.org/x/crypto/bcrypt
  ```

### 3.2 Project Structure
- [ ] Create directory structure:
  ```
  backend/
  ├── cmd/
  │   └── server/
  │       └── main.go
  ├── internal/
  │   ├── config/       # Configuration and env loading
  │   ├── database/     # PostgreSQL connection
  │   ├── models/       # User model
  │   ├── auth/         # Auth handlers and JWT
  │   ├── middleware/   # Auth middleware
  │   └── utils/        # Helper functions
  ├── migrations/
  ├── .env
  └── Dockerfile
  ```

### 3.3 Configuration & Database Connection
- [ ] Create `internal/config/config.go` - load environment variables
- [ ] Create `internal/database/postgres.go` - PostgreSQL connection with GORM
- [ ] Create `internal/models/user.go` - User model
- [ ] Implement auto-migration on startup

### 3.4 Authentication Implementation
- [ ] Create `internal/auth/handlers.go`:
  - Register handler (hash password, save to PostgreSQL)
  - Login handler (verify password, generate JWT)
  - Logout handler (optional - can be client-side only)
- [ ] Create `internal/auth/jwt.go` - JWT generation and validation
- [ ] Create `internal/middleware/auth.go` - JWT authentication middleware

### 3.5 API Routes Setup
- [ ] Create `cmd/server/main.go` with Gin router:
  ```
  POST /api/auth/register
  POST /api/auth/login
  POST /api/auth/logout
  ```
- [ ] Add CORS middleware
- [ ] Add request logging middleware

### 3.6 Testing Authentication
- [ ] Test registration endpoint with Postman/curl
- [ ] Test login endpoint and verify JWT
- [ ] Test protected routes with middleware

**Deliverables**: Working authentication API with PostgreSQL

---

## Phase 4: Backend - TigerBeetle Integration (4-5 hours)

### 4.1 TigerBeetle Client Setup
- [ ] Install TigerBeetle Go client:
  ```bash
  go get github.com/tigerbeetle/tigerbeetle-go
  ```
- [ ] Create `internal/tigerbeetle/client.go` - TigerBeetle client wrapper
- [ ] Implement connection and health check

### 4.2 Account Management
- [ ] Create `internal/account/service.go`:
  - CreateAccount() - create TigerBeetle account for new user
  - GetBalance() - query account balance from TigerBeetle
  - GetAccountInfo() - get account details
- [ ] Create `internal/account/handlers.go`:
  ```
  GET /api/accounts/me
  GET /api/accounts/balance
  ```
- [ ] Update registration flow to create TigerBeetle account
- [ ] Store TigerBeetle account ID in PostgreSQL users table

### 4.3 Transaction Operations
- [ ] Study TigerBeetle transfer model (pending, posted, flags)
- [ ] Create `internal/transaction/service.go`:
  - Deposit() - transfer from system account to user account
  - Withdraw() - transfer from user account to system account (validate balance)
  - Transfer() - transfer between user accounts (validate both accounts)
  - GetHistory() - query transfers involving user account
- [ ] Create `internal/transaction/handlers.go`:
  ```
  POST /api/transactions/deposit
  POST /api/transactions/withdraw
  POST /api/transactions/transfer
  GET /api/transactions/history?page=1&limit=10
  ```

### 4.4 Transaction Validation & Error Handling
- [ ] Validate account ownership before operations
- [ ] Handle insufficient funds error
- [ ] Handle invalid account errors
- [ ] Implement proper error responses
- [ ] Add transaction amount validation (positive, reasonable limits)

### 4.5 Testing Financial Operations
- [ ] Test deposit operation
- [ ] Test withdrawal with sufficient/insufficient funds
- [ ] Test transfer between two accounts
- [ ] Test transaction history retrieval
- [ ] Verify balance consistency after multiple operations

**Deliverables**: Complete financial operations API with TigerBeetle

---

## Phase 5: Backend - AI Chat Integration (3-4 hours)

### 5.1 MCP SDK Setup
- [ ] Research Model Context Protocol Go SDK
- [ ] Install MCP dependencies:
  ```bash
  go get github.com/modelcontextprotocol/go-sdk
  ```
- [ ] Setup OpenRouter account and get API key
- [ ] Add MCP/OpenRouter credentials to .env

### 5.2 Chat Service Implementation
- [ ] Create `internal/chat/client.go` - MCP client wrapper
- [ ] Create `internal/chat/intents.go` - Intent parsing logic
- [ ] Define intents:
  - CHECK_BALANCE
  - DEPOSIT
  - WITHDRAW
  - TRANSFER
  - VIEW_HISTORY
- [ ] Implement intent extraction from user messages

### 5.3 Chat Handler & Integration
- [ ] Create `internal/chat/service.go`:
  - ParseIntent() - extract intent and parameters from message
  - ExecuteIntent() - call appropriate banking operation
  - FormatResponse() - create user-friendly response
- [ ] Create `internal/chat/handlers.go`:
  ```
  POST /api/chat
  ```
- [ ] Implement confirmation flow for critical operations (transfers, withdrawals)

### 5.4 Chat Context Management
- [ ] Store conversation context (optional - can be stateless)
- [ ] Handle multi-turn conversations for confirmations
- [ ] Add safety checks before executing operations

### 5.5 Testing AI Chat
- [ ] Test balance inquiry via chat
- [ ] Test deposit via chat
- [ ] Test transfer via chat with confirmation
- [ ] Test error handling for ambiguous requests

**Deliverables**: Working AI chat API with MCP integration

---

## Phase 6: Backend - Docker & Finalization (1-2 hours)

### 6.1 Backend Dockerfile
- [ ] Create `backend/Dockerfile`:
  - Multi-stage build (build + production)
  - Install dependencies
  - Build Go binary
  - Expose port 8080
- [ ] Test Docker build locally

### 6.2 Update Docker Compose
- [ ] Add backend service to docker-compose.yml
- [ ] Configure environment variables
- [ ] Setup health checks
- [ ] Configure service dependencies (wait for databases)

### 6.3 Data Seeding
- [ ] Create seed script to load users from `datos-prueba-HNL.json`
- [ ] Create initial TigerBeetle accounts for seed users
- [ ] Add sample transactions for testing

### 6.4 Backend Testing
- [ ] Test full backend with `docker-compose up`
- [ ] Verify all endpoints work within Docker
- [ ] Test database persistence after restart

**Deliverables**: Dockerized backend integrated with databases

---

## Phase 7: Frontend - Setup & Authentication (2-3 hours)

### 7.1 React Project Initialization
- [ ] Navigate to `frontend/` directory
- [ ] Create Vite + React project:
  ```bash
  npm create vite@latest . -- --template react
  ```
- [ ] Install dependencies:
  ```bash
  npm install react-router-dom axios
  npm install -D tailwindcss postcss autoprefixer
  npx tailwindcss init -p
  ```

### 7.2 Project Structure
- [ ] Setup directory structure:
  ```
  frontend/
  ├── src/
  │   ├── components/
  │   │   ├── Layout.jsx
  │   │   ├── Navbar.jsx
  │   │   └── ProtectedRoute.jsx
  │   ├── pages/
  │   │   ├── Login.jsx
  │   │   ├── Register.jsx
  │   │   ├── Dashboard.jsx
  │   │   ├── Transactions.jsx
  │   │   └── History.jsx
  │   ├── services/
  │   │   └── api.js
  │   ├── context/
  │   │   └── AuthContext.jsx
  │   ├── App.jsx
  │   └── main.jsx
  ├── .env
  └── Dockerfile
  ```

### 7.3 TailwindCSS Configuration
- [ ] Configure tailwind.config.js
- [ ] Add Tailwind directives to index.css
- [ ] Test Tailwind classes

### 7.4 API Service & Auth Context
- [ ] Create `services/api.js`:
  - Axios instance with base URL
  - Request interceptor (add JWT token)
  - Response interceptor (handle 401 errors)
- [ ] Create `context/AuthContext.jsx`:
  - Login function
  - Logout function
  - User state management
  - Token storage (localStorage)

### 7.5 Authentication Pages
- [ ] Create `pages/Login.jsx`:
  - Email/password form
  - Form validation
  - Error handling
  - Redirect to dashboard on success
- [ ] Create `pages/Register.jsx`:
  - Registration form (email, password, full name)
  - Form validation
  - Success/error messages
  - Auto-login after registration

### 7.6 Routing & Protected Routes
- [ ] Setup React Router in App.jsx
- [ ] Create ProtectedRoute component
- [ ] Define routes:
  ```
  /login
  /register
  /dashboard (protected)
  /transactions (protected)
  /history (protected)
  ```

**Deliverables**: Working authentication UI

---

## Phase 8: Frontend - Banking Features (3-4 hours)

### 8.1 Dashboard Page
- [ ] Create `pages/Dashboard.jsx`:
  - Display user info and account balance
  - Show recent transactions (last 5)
  - Quick action buttons (deposit, withdraw, transfer)
  - Placeholder for chat component
- [ ] Fetch balance on component mount
- [ ] Fetch recent transactions
- [ ] Add loading states

### 8.2 Transactions Page
- [ ] Create `pages/Transactions.jsx`:
  - Tabs or sections for: Deposit, Withdraw, Transfer
  - Deposit form (amount)
  - Withdraw form (amount, balance validation)
  - Transfer form (destination account, amount)
  - Form validation
  - Success/error messages
  - Update balance after transaction

### 8.3 History Page
- [ ] Create `pages/History.jsx`:
  - Display transaction history in table/list
  - Show: date, type, amount, account involved, status
  - Implement pagination
  - Add filters (optional): date range, type
  - Loading and empty states

### 8.4 Components & UI Polish
- [ ] Create `components/Navbar.jsx`:
  - Navigation links
  - User info
  - Logout button
- [ ] Create `components/Layout.jsx` - common layout wrapper
- [ ] Add loading spinners
- [ ] Add error message components
- [ ] Improve responsive design

**Deliverables**: Complete banking UI without chat

---

## Phase 9: Frontend - AI Chat Integration (2-3 hours)

### 9.1 Chat Component
- [ ] Create `components/ChatWidget.jsx`:
  - Chat message list (scrollable)
  - Input field for user messages
  - Send button
  - Loading indicator during API call
  - Display user messages and AI responses
  - Message timestamps

### 9.2 Chat Service
- [ ] Create `services/chat.js`:
  - sendMessage() - POST to /api/chat
  - Handle streaming responses (optional)
- [ ] Add chat state management (messages array)

### 9.3 Chat Integration in Dashboard
- [ ] Integrate ChatWidget in Dashboard
- [ ] Position chat (sidebar or floating widget)
- [ ] Add toggle to show/hide chat
- [ ] Handle chat responses that trigger UI updates (balance change, etc.)

### 9.4 Confirmation Flow
- [ ] Handle AI responses requiring confirmation
- [ ] Add confirm/cancel buttons for critical operations
- [ ] Execute operation only after user confirmation
- [ ] Display operation results in chat

### 9.5 Chat Testing
- [ ] Test balance inquiry
- [ ] Test deposit via chat
- [ ] Test transfer via chat
- [ ] Test error handling

**Deliverables**: Working AI chat interface

---

## Phase 10: Frontend - Docker & Integration (1-2 hours)

### 10.1 Frontend Dockerfile
- [ ] Create `frontend/Dockerfile`:
  - Multi-stage build
  - Install dependencies
  - Build production bundle
  - Serve with nginx or Node.js
- [ ] Create nginx.conf (if using nginx)

### 10.2 Update Docker Compose
- [ ] Add frontend service to docker-compose.yml
- [ ] Configure environment variables (API URL)
- [ ] Expose port 3000 or 5173
- [ ] Setup service dependencies

### 10.3 Environment Configuration
- [ ] Create .env for frontend with API base URL
- [ ] Configure Vite to use environment variables
- [ ] Test different environments (dev vs production)

**Deliverables**: Dockerized frontend

---

## Phase 11: Testing & Documentation (2-3 hours)

### 11.1 End-to-End Testing
- [ ] Test complete user flow with Docker:
  - Register new user
  - Login
  - View dashboard and balance
  - Perform deposit
  - Perform withdrawal
  - Perform transfer to another user
  - View transaction history
  - Use chat to perform operations
- [ ] Test with users from datos-prueba-HNL.json
- [ ] Test error scenarios

### 11.2 Documentation
- [ ] Update README.md with:
  - Project description
  - Tech stack
  - Architecture diagram (optional)
  - Setup instructions (Docker)
  - Environment variables
  - API endpoints documentation
  - Test credentials
- [ ] Document architectural decisions
- [ ] Add code comments for complex logic
- [ ] Create API documentation (optional: Swagger)

### 11.3 Code Cleanup
- [ ] Remove console.logs and debug code
- [ ] Format code consistently
- [ ] Check for security issues
- [ ] Remove unused imports and dependencies

### 11.4 Final Testing
- [ ] Fresh clone of repository
- [ ] Test `docker-compose up` from scratch
- [ ] Verify all features work
- [ ] Test on different browser

**Deliverables**: Complete, documented, working system

---

## Phase 12: Deployment & Submission (1 hour)

### 12.1 Repository Finalization
- [ ] Clean git history
- [ ] Write meaningful commit messages
- [ ] Tag release version (v1.0.0)
- [ ] Verify all files are committed
- [ ] Check .gitignore is working (.env excluded)

### 12.2 Optional: Deployment
- [ ] Deploy to Railway/Render/Vercel (if time permits)
- [ ] Update README with live demo URL

### 12.3 Submission
- [ ] Verify README has all required information
- [ ] Test GitHub repository is public
- [ ] Prepare demonstration
- [ ] Submit repository URL

**Deliverables**: Submitted project

---

## Timeline Summary

| Phase | Duration | Cumulative |
|-------|----------|------------|
| 1. Project Setup | 0.5 hours | 0.5 hours |
| 2. Database Infrastructure | 2-3 hours | 3.5 hours |
| 3. Backend Auth | 3-4 hours | 7.5 hours |
| 4. Backend TigerBeetle | 4-5 hours | 12.5 hours |
| 5. Backend AI Chat | 3-4 hours | 16.5 hours |
| 6. Backend Docker | 1-2 hours | 18.5 hours |
| 7. Frontend Auth | 2-3 hours | 21.5 hours |
| 8. Frontend Banking | 3-4 hours | 25.5 hours |
| 9. Frontend Chat | 2-3 hours | 28.5 hours |
| 10. Frontend Docker | 1-2 hours | 30.5 hours |
| 11. Testing & Docs | 2-3 hours | 33.5 hours |
| 12. Submission | 1 hour | 34.5 hours |

**Total Estimated Time**: 30-35 hours
**Available Time**: 48 hours (2 days)
**Buffer**: ~13-15 hours for debugging and issues

---

## Priority Levels

### Must Have (P0) - Core Requirements
- Authentication (register, login)
- Account management (balance)
- Transactions (deposit, withdraw, transfer)
- Transaction history
- AI chat integration with MCP
- Docker setup
- Documentation

### Should Have (P1) - Important but Optional
- Pagination for history
- Form validations
- Error handling
- Loading states
- Responsive design

### Nice to Have (P2) - Bonus Features
- Tests (unit, integration)
- CI/CD with GitHub Actions
- Rate limiting
- Structured logging
- Data visualizations
- Export to PDF/CSV

---

## Risk Mitigation

### High Risk Areas
1. **TigerBeetle Integration**: Unfamiliar technology, complex accounting model
   - Mitigation: Study docs first, test extensively, keep implementation simple

2. **MCP Integration**: New protocol, potential compatibility issues
   - Mitigation: Start with simple chat, test early, have fallback plan

3. **Docker Configuration**: All services must work together
   - Mitigation: Test each service individually first, use health checks

### Contingency Plans
- If TigerBeetle is too complex: Simplify to basic account balance tracking
- If MCP doesn't work: Use direct API calls to OpenAI/Anthropic
- If time runs short: Focus on P0 features, skip P2 features

---

## Next Steps

**START HERE**:

1. Create GitHub repository
2. Setup initial project structure
3. Create docker-compose.yml with databases
4. Begin backend authentication

**Remember**: Commit frequently with clear messages!
