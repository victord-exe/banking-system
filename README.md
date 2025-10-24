# Online Banking System

A full-stack online banking application with AI-powered chat interface built with Go, React, TigerBeetle, and Model Context Protocol (MCP).

## Features

- User authentication (register, login, JWT)
- Account management with real-time balance
- Financial transactions (deposit, withdraw, transfer)
- Transaction history with pagination
- AI-powered chat interface for natural language banking operations
- Dual database architecture (TigerBeetle for financial data, PostgreSQL for user data)

## Tech Stack

### Backend
- Go 1.21+
- Gin (HTTP framework)
- GORM (PostgreSQL ORM)
- TigerBeetle (financial database)
- JWT authentication
- Model Context Protocol (MCP) for AI integration

### Frontend
- React 18+ with Vite
- React Router
- Axios
- TailwindCSS
- AI Chat Interface

### Infrastructure
- Docker & Docker Compose
- PostgreSQL
- TigerBeetle

## Quick Start

### Prerequisites
- Docker and Docker Compose installed
- Git

### Running the Application

1. Clone the repository:
```bash
git clone <repository-url>
cd HLABS
```

2. Create environment file:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Start all services:
```bash
docker-compose up
```

4. Access the application:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

## Development

See [CLAUDE.md](CLAUDE.md) and [WORKPLAN.md](WORKPLAN.md) for detailed development guidelines and project roadmap.

### Backend Development
```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

### Frontend Development
```bash
cd frontend
npm install
npm run dev
```

## Architecture

### Dual Database Design
- **TigerBeetle**: Handles all financial operations (accounts, transactions, balances)
- **PostgreSQL**: Manages user authentication and profile data

### API Endpoints

#### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user
- `POST /api/auth/logout` - Logout user

#### Accounts
- `GET /api/accounts/me` - Get account info
- `GET /api/accounts/balance` - Get current balance

#### Transactions
- `POST /api/transactions/deposit` - Deposit funds
- `POST /api/transactions/withdraw` - Withdraw funds
- `POST /api/transactions/transfer` - Transfer to another account
- `GET /api/transactions/history` - Get transaction history

#### AI Chat
- `POST /api/chat` - Send message to AI assistant

## Test Data

Test users are provided in `datos-prueba-HNL.json`. All test passwords follow the pattern: `{FirstName}2024!`

Example:
- Email: `ihernandez@email.com`
- Password: `Isabel2024!`

## Environment Variables

Create a `.env` file in the root directory:

```env
# Backend
PORT=8080
JWT_SECRET=your-secret-key-here
DATABASE_URL=postgresql://user:password@localhost:5432/banking

# TigerBeetle
TIGERBEETLE_HOST=localhost
TIGERBEETLE_PORT=3000

# AI/MCP
OPENROUTER_API_KEY=your-openrouter-key
MCP_MODEL=anthropic/claude-3-5-sonnet

# Frontend
VITE_API_URL=http://localhost:8080
```

## Project Structure

```
HLABS/
├── backend/
│   ├── cmd/
│   │   └── server/         # Application entry point
│   ├── internal/
│   │   ├── auth/          # Authentication logic
│   │   ├── account/       # Account management
│   │   ├── transaction/   # Transaction handlers
│   │   ├── chat/          # AI/MCP integration
│   │   ├── tigerbeetle/   # TigerBeetle client
│   │   └── database/      # PostgreSQL client
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/    # Reusable components
│   │   ├── pages/         # Page components
│   │   ├── services/      # API clients
│   │   └── context/       # React context
│   └── Dockerfile
├── docker-compose.yml
├── .env.example
└── README.md
```

## License

This project was created as a technical test.

## Contact

For questions or issues, please open an issue in the repository.
