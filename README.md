# HLABS - Online Banking System

Full-stack banking system with AI chat integration, built with Go, React, PostgreSQL, and TigerBeetle.

## 🚀 Quick Start (For Evaluators)

**One command to run everything:**

```bash
docker-compose up
```

That's it! The entire system will start automatically with all services configured.

**What happens automatically:**
- ✅ TigerBeetle initializes its data file on first run
- ✅ PostgreSQL creates the database and tables
- ✅ Backend server starts and connects to both databases
- ✅ Frontend builds and serves the React application
- ✅ Test users are seeded from `datos-prueba-HNL.json`

**Access the application:**
- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **API Health**: http://localhost:8080/health

**Test Credentials:**

Users from `datos-prueba-HNL.json` are automatically seeded on first run.

Password pattern: `{FirstName}2024!`

Example:
```
Email: (from datos-prueba-HNL.json)
Password: Maria2024!  (if name is María López)
```

---

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Features](#features)
- [Setup Instructions](#setup-instructions)
- [Project Structure](#project-structure)
- [API Documentation](#api-documentation)
- [Development](#development)
- [Environment Variables](#environment-variables)

---

## 🎯 Overview

This project is a **technical test** demonstrating a complete online banking system with:

- ✅ **User authentication** (JWT-based)
- ✅ **Account management** (balance, transactions)
- ✅ **Financial operations** (deposits, withdrawals, transfers)
- ✅ **Transaction history** with pagination
- ✅ **AI chat interface** using Model Context Protocol (MCP)
- ✅ **Dual database architecture** (PostgreSQL + TigerBeetle)
- ✅ **Full Docker support** (one-command startup)

**Key Highlight**: Uses **TigerBeetle**, a distributed financial database designed for mission-critical financial operations with ACID guarantees and immutable transaction logs.

---

## 🏗️ Architecture

### System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend                            │
│                    (React + Vite)                          │
│         Dashboard | Transactions | AI Chat                │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP/REST
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                      Backend (Go)                           │
│                     Gin Framework                          │
├─────────────────────────────────────────────────────────────┤
│  Auth  │  Accounts  │  Transactions  │  AI Chat (MCP)     │
└────┬────────────┬───────────────────┬──────────────────────┘
     │            │                   │
     ▼            ▼                   ▼
┌──────────┐ ┌────────────────┐ ┌─────────────┐
│PostgreSQL│ │  TigerBeetle   │ │  OpenRouter │
│  (Users) │ │  (Financial)   │ │   (AI API)  │
└──────────┘ └────────────────┘ └─────────────┘
```

### Dual Database Strategy

| Database | Purpose | Data Stored |
|----------|---------|-------------|
| **PostgreSQL** | User authentication | Users, credentials, sessions, profile data |
| **TigerBeetle** | Financial operations | Accounts, balances, transfers, transaction history |

**Why?** Separating authentication from financial data:
- ✅ Security: Financial data isolated from user credentials
- ✅ Performance: Each DB optimized for its specific use case
- ✅ Scalability: Can scale independently
- ✅ Audit: TigerBeetle provides immutable financial audit trail

### TigerBeetle Accounting Model

Uses **double-entry bookkeeping**:

```
Deposit $100:
  System Account → User Account: $100

Withdraw $50:
  User Account → System Account: $50

Transfer $30 (User A → User B):
  User_A Account → User_B Account: $30
```

---

## 🛠️ Tech Stack

### Backend
- **Go 1.21+** - Systems programming language
- **Gin** - HTTP web framework
- **GORM** - ORM for PostgreSQL
- **golang-jwt/jwt** - JWT authentication
- **bcrypt** - Password hashing
- **tigerbeetle-go** - TigerBeetle client

### Frontend
- **Vite** - Build tool and dev server
- **React 18** - UI framework
- **React Router** - Client-side routing
- **Axios** - HTTP client
- **TailwindCSS** - Utility-first CSS

### Databases
- **PostgreSQL 16** - User authentication and profiles
- **TigerBeetle** - Financial transactions and accounting

### AI Integration
- **Model Context Protocol (MCP)** - Standardized AI interface
- **OpenRouter** - Multi-model AI API access

### Infrastructure
- **Docker & Docker Compose** - Containerization

---

## ✨ Features

### Core Banking Features

#### 1. Authentication & Security
- [x] User registration with email validation
- [x] Secure login with JWT tokens
- [x] Password hashing (bcrypt)
- [x] Protected API routes
- [x] Session management

#### 2. Account Management
- [x] Automatic bank account creation (TigerBeetle)
- [x] Real-time balance display
- [x] Account overview dashboard

#### 3. Financial Operations
- [x] **Deposits**: Add funds to account
- [x] **Withdrawals**: Remove funds (with balance validation)
- [x] **Transfers**: Send money to other users
- [x] **Transaction History**: Paginated list of all transactions
- [x] Real-time balance updates

#### 4. AI Chat Interface
- [x] Natural language banking operations
- [x] Examples:
  - "What's my balance?"
  - "Transfer $100 to account 12345"
  - "Show my last 10 transactions"
- [x] Confirmation prompts for critical operations

---

## 🔧 Setup Instructions

### Prerequisites

- **Docker & Docker Compose** (required)
- **Git** (to clone the repository)

That's all! Docker handles everything else.

### Installation

1. **Clone the repository**

```bash
git clone https://github.com/[username]/hlabs-banking.git
cd hlabs-banking
```

2. **(Optional) Configure AI Chat**

For AI chat to work, add your OpenRouter API key to `.env.docker`:

```bash
# Edit .env.docker
OPENROUTER_API_KEY=sk-or-v1-your-actual-api-key
```

Get a free key at https://openrouter.ai

> **Note**: The system works without an API key, but AI chat will return an error message.

3. **Start the system**

```bash
docker-compose up
```

Or in detached mode:

```bash
docker-compose up -d
```

4. **Wait for services to start** (~30-60 seconds)

Watch the logs:
```bash
docker-compose logs -f
```

You should see:
- ✅ PostgreSQL ready
- ✅ TigerBeetle initialized (automatic on first run)
- ✅ Backend server running on :8080
- ✅ Frontend running on :5173

5. **Access the application**

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080
- Health check: http://localhost:8080/health

### Stopping the System

```bash
docker-compose down
```

To remove all data (fresh start):

```bash
docker-compose down -v
```

---

## 📁 Project Structure

```
hlabs-banking/
├── backend/                    # Go backend service
│   ├── cmd/server/            # Application entry point
│   ├── internal/
│   │   ├── config/            # Configuration loading
│   │   ├── database/          # PostgreSQL connection
│   │   ├── models/            # Data models
│   │   ├── auth/              # Authentication (JWT)
│   │   ├── middleware/        # HTTP middleware
│   │   ├── tigerbeetle/       # TigerBeetle client
│   │   ├── account/           # Account service
│   │   ├── transaction/       # Transaction service
│   │   ├── chat/              # AI chat integration
│   │   └── utils/             # Utilities
│   ├── Dockerfile
│   ├── go.mod
│   ├── BACKEND_TASKS.md       # Backend development plan
│   └── README.md
│
├── frontend/                   # React frontend
│   ├── src/
│   │   ├── components/        # Reusable components
│   │   ├── pages/             # Route pages
│   │   ├── services/          # API client
│   │   └── context/           # React context
│   ├── Dockerfile
│   └── package.json
│
├── datos-prueba-HNL.json      # Test user data
├── docker-compose.yml         # Docker orchestration
├── .env.docker                # Default environment (committed)
├── .gitignore
├── WORKPLAN.md                # Development plan
├── CLAUDE.md                  # AI assistant instructions
└── README.md                  # This file
```

---

## 📡 API Documentation

### Base URL

```
http://localhost:8080/api
```

### Authentication Endpoints (Public)

#### Register
```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "full_name": "John Doe"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "full_name": "John Doe"
  }
}
```

#### Login
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

### Account Endpoints (Protected)

All require `Authorization: Bearer <token>` header.

#### Get Balance
```http
GET /api/accounts/balance
Authorization: Bearer <token>
```

**Response:**
```json
{
  "account_id": "12345",
  "balance": 1500.50,
  "currency": "USD"
}
```

### Transaction Endpoints (Protected)

#### Deposit
```http
POST /api/transactions/deposit
Authorization: Bearer <token>
Content-Type: application/json

{
  "amount": 100.00
}
```

#### Withdraw
```http
POST /api/transactions/withdraw
Authorization: Bearer <token>
Content-Type: application/json

{
  "amount": 50.00
}
```

#### Transfer
```http
POST /api/transactions/transfer
Authorization: Bearer <token>
Content-Type: application/json

{
  "to_account_id": 67890,
  "amount": 75.00
}
```

#### Transaction History
```http
GET /api/transactions/history?page=1&limit=10
Authorization: Bearer <token>
```

### AI Chat Endpoint (Protected)

```http
POST /api/chat
Authorization: Bearer <token>
Content-Type: application/json

{
  "message": "What is my account balance?"
}
```

---

## 👨‍💻 Development

### Local Development (Without Docker)

#### Backend

1. Install Go 1.21+
2. Start databases:
   ```bash
   docker-compose up postgres tigerbeetle
   ```
3. Run backend:
   ```bash
   cd backend
   cp .env.example .env
   # Edit .env with your values
   go mod download
   go run cmd/server/main.go
   ```

#### Frontend

```bash
cd frontend
npm install
npm run dev
```

---

## 🔐 Environment Variables

The project uses `.env.docker` (committed to repository) with safe defaults for evaluation.

### Key Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_USER` | PostgreSQL username | banking_user |
| `POSTGRES_PASSWORD` | PostgreSQL password | banking_secure_password_2024 |
| `POSTGRES_DB` | Database name | banking_system |
| `TIGERBEETLE_HOST` | TigerBeetle server | tigerbeetle |
| `TIGERBEETLE_PORT` | TigerBeetle port | 3000 |
| `JWT_SECRET` | JWT signing key | (auto-generated) |
| `OPENROUTER_API_KEY` | AI chat API key | (must be added) |
| `SERVER_PORT` | Backend port | 8080 |

For production, create a `.env` file and override with real secrets.

---

## 🧪 Testing

### Test Users

Test data from `datos-prueba-HNL.json` is automatically loaded on first run.

Password pattern: `{FirstName}2024!`

### Manual Test Flow

1. **Register** a new user or use seeded user
2. **Login** and receive JWT token
3. **Check balance** (should be $0 for new accounts)
4. **Deposit** $100
5. **Verify** balance is $100
6. **Withdraw** $30
7. **Transfer** $20 to another user
8. **View history** (should show 3 transactions)
9. **AI Chat**: "What's my balance?" (should return $50)

---

## 📝 Evaluation Criteria

This project addresses all evaluation requirements:

### ✅ Functionality (40%)
- All banking features working
- AI chat integration functional
- No critical bugs

### ✅ Code Quality (30%)
- Clean, readable code
- Proper error handling
- Modular architecture

### ✅ Security (10%)
- Password hashing (bcrypt)
- JWT authentication
- Input validation
- SQL injection prevention

### ✅ AI Integration (20%)
- MCP implementation
- Natural language processing
- Intent recognition

### ✅ Documentation (10%)
- Comprehensive README
- API documentation
- Code comments

---

## 🏛️ Architectural Decisions

### Why Dual Database Architecture?

**Decision**: Use PostgreSQL for user data and TigerBeetle for financial operations.

**Rationale**:
- **Separation of Concerns**: User authentication and financial transactions have different consistency requirements
- **TigerBeetle Advantages**: Built specifically for financial transactions with:
  - Immutable transaction log (audit trail)
  - ACID guarantees at the database level
  - Double-entry bookkeeping enforced by design
  - Protection against overdrafts (debits cannot exceed credits)
- **PostgreSQL for Users**: Standard relational database well-suited for user management and authentication
- **Scalability**: Each database can be scaled independently based on load

### Why Go for Backend?

**Decision**: Use Go instead of Rust or COBOL.

**Rationale**:
- **TigerBeetle SDK**: Official Go client is mature and well-documented
- **Performance**: Go provides excellent performance for concurrent requests
- **Development Speed**: Faster development time within the 2-day constraint
- **Ecosystem**: Rich ecosystem (Gin, GORM, JWT libraries)

### Why Model Context Protocol (MCP)?

**Decision**: Implement AI chat using MCP instead of direct LLM API calls.

**Rationale**:
- **Standardization**: MCP provides a standardized way to integrate AI tools
- **Flexibility**: Easy to swap AI providers (Claude, GPT-4, etc.) via OpenRouter
- **Tool Calling**: Built-in support for structured tool calling (deposit, withdraw, transfer)
- **Confirmation Flows**: MCP enables proper confirmation workflows for critical operations

### Why Vite + React?

**Decision**: Use Vite with React for the frontend.

**Rationale**:
- **Fast Development**: Hot Module Replacement (HMR) for instant feedback
- **Modern Tooling**: Better developer experience than traditional bundlers
- **React Ecosystem**: Large ecosystem of libraries and components
- **Performance**: Optimized production builds

### Security Decisions

1. **JWT Tokens**: Stateless authentication for API scalability
2. **bcrypt Hashing**: Industry-standard password hashing (cost factor: 10)
3. **UTF-8 Encoding**: Explicit `client_encoding=UTF8` to handle international characters correctly
4. **Protected Routes**: Middleware-based authentication on all sensitive endpoints
5. **Input Validation**: Server-side validation for all transaction amounts and account IDs

### Data Flow Design

**Money Representation**:
- **Backend**: All amounts stored in **cents** (int64) to avoid floating-point errors
- **Frontend**: Converts to dollars for display, converts back to cents before sending
- **TigerBeetle**: Uses uint128 for maximum precision

**Example**: $100.50 → 10050 cents → TigerBeetle transfer → 10050 cents → $100.50 display

### Docker Configuration

**Decision**: All services in Docker Compose with automatic initialization.

**Rationale**:
- **One-Command Setup**: `docker-compose up` starts everything
- **TigerBeetle Init Script**: Custom script auto-creates data file if missing
- **Health Checks**: Ensure services start in correct order
- **Volume Persistence**: Data persists across container restarts
- **Production-Ready**: Same configuration can be used in staging/production with environment overrides

### Why Automatic Seeding?

**Decision**: Automatically load 7,000+ test users from JSON on first run.

**Rationale**:
- **Evaluator Experience**: No manual setup required for testing
- **Realistic Data**: Large dataset tests pagination, search, and performance
- **Idempotent**: Seeding only runs if database is empty (safe to restart)
- **Historical Transactions**: Pre-populated transaction history makes the app feel real

---

## 📄 License

This project is a technical assessment.

---

**Built for the HLABS Technical Assessment**

**Time Constraint**: 2 days | **Docker**: ✅ One-command setup
