# HLABS Banking System - Backend

Go-based backend for the online banking system with TigerBeetle integration and AI chat capabilities.

## Tech Stack

- **Go 1.21+** - Programming language
- **Gin** - HTTP web framework
- **GORM** - ORM for PostgreSQL
- **PostgreSQL** - User authentication and data
- **TigerBeetle** - Distributed financial database
- **JWT** - Authentication tokens
- **MCP (Model Context Protocol)** - AI chat integration
- **OpenRouter** - AI model access

## Project Structure

```
backend/
├── cmd/
│   └── server/           # Main application entry point
│       └── main.go
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # PostgreSQL connection
│   ├── models/          # Data models (User)
│   ├── auth/            # Authentication (JWT, handlers)
│   ├── middleware/      # HTTP middleware (auth)
│   ├── tigerbeetle/     # TigerBeetle client wrapper
│   ├── account/         # Account management service
│   ├── transaction/     # Transaction operations
│   ├── chat/            # AI chat integration (MCP)
│   └── utils/           # Utility functions
├── migrations/          # Database migrations
├── scripts/             # Utility scripts (seeding, etc.)
├── .env                 # Environment variables (not in git)
├── .env.example         # Example environment variables
├── go.mod               # Go module definition
└── Dockerfile           # Docker image definition
```

## Prerequisites

Before running the backend, ensure you have:

1. **Go 1.21 or higher** installed
   - Download: https://go.dev/doc/install
   - Verify: `go version`

2. **PostgreSQL** running (via Docker or local install)

3. **TigerBeetle** running (via Docker)

4. **OpenRouter API key** (for AI chat)
   - Sign up at https://openrouter.ai

## Setup Instructions

### 1. Install Dependencies

```bash
cd backend
go mod download
```

### 2. Configure Environment Variables

```bash
# Copy example env file
cp .env.example .env

# Edit .env with your actual values
# Required: JWT_SECRET, OPENROUTER_API_KEY, database credentials
```

### 3. Run Locally (Development)

Make sure PostgreSQL and TigerBeetle are running first:

```bash
# Start databases (from root directory)
docker-compose up postgres tigerbeetle

# Run backend
go run cmd/server/main.go
```

Server will start on `http://localhost:8080`

### 4. Run with Docker

From the root directory:

```bash
docker-compose up backend
```

## API Endpoints

### Authentication (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register new user |
| POST | `/api/auth/login` | Login and get JWT token |
| POST | `/api/auth/logout` | Logout (client-side) |

### Accounts (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/accounts/me` | Get current user's account |
| GET | `/api/accounts/balance` | Get account balance |

### Transactions (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/transactions/deposit` | Deposit funds |
| POST | `/api/transactions/withdraw` | Withdraw funds |
| POST | `/api/transactions/transfer` | Transfer to another account |
| GET | `/api/transactions/history` | Get transaction history |

### AI Chat (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/chat` | Send message to AI assistant |

## Example Requests

### Register

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!",
    "full_name": "John Doe"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'
```

### Get Balance (with JWT)

```bash
curl -X GET http://localhost:8080/api/accounts/balance \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Deposit

```bash
curl -X POST http://localhost:8080/api/transactions/deposit \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100.00
  }'
```

### Transfer

```bash
curl -X POST http://localhost:8080/api/transactions/transfer \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "to_account_id": 12345,
    "amount": 50.00
  }'
```

### AI Chat

```bash
curl -X POST http://localhost:8080/api/chat \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What is my account balance?"
  }'
```

## Environment Variables

See [.env.example](.env.example) for all available configuration options.

Key variables:

- `SERVER_PORT` - HTTP server port (default: 8080)
- `POSTGRES_DSN` - PostgreSQL connection string
- `TIGERBEETLE_HOST` - TigerBeetle server address
- `JWT_SECRET` - Secret key for JWT signing
- `OPENROUTER_API_KEY` - API key for AI chat

## Development Workflow

### 1. Run Tests

```bash
go test ./...
```

### 2. Build Binary

```bash
go build -o bin/server cmd/server/main.go
./bin/server
```

### 3. Seed Test Data

```bash
go run scripts/seed.go
```

This will create test users from `datos-prueba-HNL.json`.

### 4. Format Code

```bash
go fmt ./...
```

### 5. Check for Issues

```bash
go vet ./...
```

## Database Architecture

### PostgreSQL (User Data)

Stores user authentication and profile information:

- **users** table: id, email, password_hash, full_name, tigerbeetle_account_id, timestamps

### TigerBeetle (Financial Data)

Stores all financial transactions using double-entry accounting:

- **Accounts**: User accounts (ledger 1), System accounts (ledger 2)
- **Transfers**: All transactions between accounts
- Immutable, audit-trail preserving

## TigerBeetle Integration

### Account Creation

When a user registers:
1. Create user record in PostgreSQL
2. Create financial account in TigerBeetle
3. Store TigerBeetle account ID in user record

### Transaction Flow

**Deposit**: System Account → User Account
**Withdraw**: User Account → System Account
**Transfer**: User Account A → User Account B

All operations are atomic and maintain consistency.

## Security Features

- **Password Hashing**: bcrypt with salt
- **JWT Authentication**: Token-based auth with expiration
- **CORS Protection**: Configured allowed origins
- **Input Validation**: Request body validation
- **SQL Injection Prevention**: GORM parameterized queries
- **Account Ownership Validation**: Users can only access their own accounts

## Troubleshooting

### "go: command not found"

Install Go from https://go.dev/doc/install

### Cannot connect to PostgreSQL

1. Check PostgreSQL is running: `docker-compose ps`
2. Verify connection string in `.env`
3. Check firewall/network settings

### Cannot connect to TigerBeetle

1. Ensure TigerBeetle container is running
2. Check `TIGERBEETLE_HOST` and `TIGERBEETLE_PORT` in `.env`
3. Review TigerBeetle logs: `docker-compose logs tigerbeetle`

### JWT token invalid

1. Check `JWT_SECRET` matches between requests
2. Verify token hasn't expired (24 hours default)
3. Ensure `Authorization: Bearer <token>` header format

## Next Steps

See [BACKEND_TASKS.md](BACKEND_TASKS.md) for detailed development tasks and progress tracking.

## License

Private project for technical assessment.
