# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 🚨 CRITICAL RULES - READ FIRST

### Command Execution Policy
**NEVER run the following commands via Bash tool unless explicitly requested by the user:**
- ❌ `go mod tidy` / `go mod download` / `go get`
- ❌ `npm install` / `npm ci` / `yarn install`
- ❌ `docker-compose up` / `docker-compose down` / `docker build`
- ❌ `go run` / `npm run` / `npm start` / any server startup commands
- ❌ Database migration commands
- ❌ Package installation or dependency management commands

**What you SHOULD do:**
- ✅ Inform the user which command they need to run
- ✅ Explain what the command does
- ✅ Provide the exact command with proper flags
- ✅ Wait for user to confirm they've run it

**Example:**
```
❌ BAD: Running `go mod tidy` to download dependencies...
✅ GOOD: "Please run `go mod tidy` in the backend directory to download Go dependencies. This command will download all required packages and clean up unused ones."
```

**Exception:** The user may explicitly ask you to run a command like: "run npm install for me" or "execute go mod tidy". Only then you may use the Bash tool.

## Project Overview

This is an **online banking system** technical test that requires building a full-stack application with AI chat integration. The system allows users to manage bank accounts, perform transactions (deposits, withdrawals, transfers), and interact with the banking system via natural language using AI.

## Technical Stack (Selected)

### Backend
- **Go 1.21+**
- **Gin** or **Chi** - HTTP web framework
- **GORM** - ORM for PostgreSQL
- **golang-jwt/jwt** - JWT authentication
- **tigerbeetle-go** - TigerBeetle client
- **bcrypt** - Password hashing

### Frontend
- **Vite** - Build tool
- **React 18+** - UI framework
- **React Router** - Client-side routing
- **Axios** - HTTP client
- **TailwindCSS** - Styling (recommended)
- **React Query** - Data fetching and caching (optional)
- **Zustand** or **Context API** - State management

### Databases (BOTH required)
- **TigerBeetle**: Distributed financial database for all financial operations (accounts, transactions, balances)
- **PostgreSQL**: User data and authentication

### Infrastructure
- **Docker & Docker Compose** (mandatory - entire system must run with `docker-compose up`)
- **GitHub** (mandatory - public repository)

### AI Integration (Mandatory)
- **Model Context Protocol (MCP)**: Using [go-sdk](https://github.com/modelcontextprotocol/go-sdk)
- **OpenRouter** or similar for AI model access (Claude, GPT-4, etc.)
- Natural language processing for banking operations

## Key Architecture Principles

### Dual Database Architecture
This project uses a **split database approach**:
- **TigerBeetle handles all financial data**: accounts, transactions, balances, transfer history
- **PostgreSQL/SQLite handles authentication**: users, credentials, session tokens

**Critical**: Each user record in PostgreSQL/SQLite must have a reference to their TigerBeetle account ID to link authentication with financial operations.

### TigerBeetle Integration Points
TigerBeetle uses a specific accounting model:
1. **Accounts**: Financial accounts with balances (not user accounts)
2. **Transfers**: All transactions are transfers between accounts
3. **Double-entry bookkeeping**: Deposits/withdrawals involve system accounts
4. **Immutability**: Transactions cannot be modified, only new correcting transactions can be created
5. **Atomic operations**: All transfers must be atomic and maintain consistency

When implementing:
- Study TigerBeetle's transfer model carefully
- Handle edge cases (insufficient funds, invalid accounts)
- Ensure all operations are properly audited
- Maintain referential integrity between PostgreSQL users and TigerBeetle accounts

## Required Commands

### Development Setup
```bash
# Start entire system
docker-compose up

# Start in detached mode
docker-compose up -d

# Stop all services
docker-compose down

# Rebuild containers
docker-compose up --build
```

### Backend Development
```bash
# Navigate to backend directory
cd backend

# Initialize Go module (first time only)
go mod init github.com/yourusername/banking-system

# Install dependencies
go mod tidy

# Run locally (requires databases running)
go run cmd/server/main.go

# Run tests
go test ./...

# Build binary
go build -o bin/server cmd/server/main.go
```

### Frontend Development
```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

### Test Data
The file `datos-prueba-HNL (1).json` contains test users with credentials:
- Format: email, password, full_name, id, created_at
- Passwords follow pattern: `{FirstName}2024!`
- Use these for seeding initial users and testing

## Core Functional Requirements

### 1. Authentication & Security
- User registration with automatic bank account creation in TigerBeetle
- Secure login with token-based authentication (JWT)
- Session management and logout
- Optional: Two-factor authentication

### 2. Account Management
- Display account balance (from TigerBeetle)
- View transaction history
- Each account must have a unique identifier linking to TigerBeetle

### 3. Transactions (via TigerBeetle)
- **Deposits**: Create transfer from system account to user account
- **Withdrawals**: Validate balance, create transfer from user account to system account
- **Transfers**: Validate destination account exists, create transfer between user accounts
- All transactions must be recorded with: timestamp, type, amount, accounts involved, status

### 4. AI Chat Interface (MANDATORY)
Integrate MCP-based AI that allows users to:
- Perform transactions: "Transfer $100 to account 12345"
- Check balance: "How much money do I have?"
- View history: "Show my last 10 transactions"
- Any banking operation in natural language

**Important**: AI must confirm critical operations before execution

### 5. Dashboard
- Account overview with balance
- Recent transactions list
- AI chat component (visible/accessible)
- Optional: Data visualizations

## API Endpoints Structure

### Authentication
- `POST /api/auth/register` - Create user + TigerBeetle account
- `POST /api/auth/login` - Authenticate and return token
- `POST /api/auth/logout` - Invalidate session

### Accounts
- `GET /api/accounts/me` - Get current user's account info
- `GET /api/accounts/balance` - Get balance from TigerBeetle

### Transactions
- `POST /api/transactions/deposit` - Add funds
- `POST /api/transactions/withdraw` - Remove funds (validate balance)
- `POST /api/transactions/transfer` - Send to another account
- `GET /api/transactions/history` - Get transaction list (with pagination)

### AI Chat
- `POST /api/chat` - Send message to AI, parse intent, execute operations

## Docker Configuration

The project must include:
- `Dockerfile` for backend service
- `Dockerfile` for frontend service
- `docker-compose.yml` orchestrating:
  - Backend container
  - Frontend container (Vite dev server or production build)
  - TigerBeetle container
  - PostgreSQL or SQLite container
  - Network configuration for inter-service communication
  - Volume mounts for data persistence
  - Environment variables configuration

**Critical**: Running `docker-compose up` must start the entire working system without additional commands.

## Code Organization Guidelines

### Backend Structure (Example for Go)
```
backend/
├── cmd/
│   └── server/          # Main application entry
├── internal/
│   ├── auth/           # Authentication logic
│   ├── account/        # Account management
│   ├── transaction/    # Transaction handlers
│   ├── chat/           # AI/MCP integration
│   ├── tigerbeetle/    # TigerBeetle client wrapper
│   └── database/       # PostgreSQL/SQLite client
├── pkg/
│   └── middleware/     # Auth middleware, logging, etc.
└── Dockerfile
```

### Frontend Structure (Example for React)
```
frontend/
├── src/
│   ├── components/     # Reusable components
│   ├── pages/          # Login, Dashboard, Transactions, History
│   ├── services/       # API client
│   ├── hooks/          # Custom React hooks
│   └── context/        # Auth context, user state
├── public/
└── Dockerfile
```

## Testing Approach

- Unit tests for transaction logic (critical for financial operations)
- Integration tests for TigerBeetle operations
- E2E tests for main user flows
- Test AI chat intent parsing separately from execution

## Security Considerations

- Hash passwords (bcrypt recommended)
- Use HTTPS in production
- Validate all inputs on backend
- Implement rate limiting on sensitive endpoints
- Never expose TigerBeetle directly to frontend
- Validate account ownership before operations
- Log all financial transactions for audit trail

## Common Pitfalls to Avoid

1. **Don't mix database responsibilities**: Keep user data separate from financial data
2. **Don't skip TigerBeetle documentation**: Its model is different from typical RDBMS
3. **Don't forget atomicity**: All multi-step financial operations must be atomic
4. **Don't skip validation**: Always validate account ownership and sufficient funds
5. **Don't ignore error handling**: Financial operations require robust error handling
6. **Don't forget Docker networking**: Services must be able to communicate within Docker network

## Resources

- TigerBeetle docs: https://docs.tigerbeetle.com/
- MCP Go SDK: https://github.com/modelcontextprotocol/go-sdk
- MCP docs: https://modelcontextprotocol.io
- OpenRouter (AI models): https://openrouter.ai

## Evaluation Criteria Summary

- **Functionality (40%)**: All features working, no critical bugs
- **Code Quality (30%)**: Clean code, good architecture, error handling
- **Security (10%)**: Security measures implemented
- **AI Integration (20%)**: MCP working, natural language processing functional
- **Documentation (10%)**: Clear README, documented endpoints and decisions

## Time Constraint

**Maximum 2 days** to complete the project.

## Deliverables Checklist

- [ ] Public GitHub repository with clear commit history
- [ ] Complete `docker-compose.yml` that starts entire system
- [ ] Backend with all required endpoints
- [ ] Frontend with Login, Dashboard, Transactions, History pages
- [ ] AI chat interface integrated with MCP
- [ ] Both databases properly configured (TigerBeetle + PostgreSQL/SQLite)
- [ ] README.md with setup instructions and architectural decisions
- [ ] SQL scripts for database initialization
- [ ] Environment variables documented
- [ ] Test credentials provided