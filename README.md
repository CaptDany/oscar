# OpenCRM

A production-grade, open-source CRM backend built in Go.

## Features

- Multi-tenant SaaS with Row Level Security
- RESTful API with cursor-based pagination
- Paseto v2 stateless authentication
- Custom fields and white-label branding
- Automation engine with triggers and actions
- WebSocket support for real-time updates
- Docker-ready for easy deployment

## Quick Start

### Prerequisites

- Go 1.23+
- Docker & Docker Compose
- Make

### Development Setup

```bash
# Clone the repository
git clone https://github.com/opencrm/opencrm.git
cd opencrm

# Copy environment file
cp .env.example .env

# Install dependencies
go mod download

# Run migrations
make migrate/up

# Start development server
make dev
```

### Docker Setup

```bash
# Start all services
make docker-up

# Stop all services
make docker-down
```

## Tech Stack

- **Language**: Go 1.23+
- **Framework**: Echo v4
- **Database**: PostgreSQL 16 with pgx/v5
- **Cache**: Redis 7
- **Queue**: Asynq
- **Auth**: Paseto v2
- **Storage**: MinIO (S3-compatible)
- **Proxy**: Caddy

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new tenant
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Refresh tokens
- `POST /api/v1/auth/logout` - Logout
- `GET /api/v1/auth/me` - Current user

### CRM
- `GET/POST /api/v1/persons` - List/Create persons
- `GET/PATCH/DELETE /api/v1/persons/:id` - Person CRUD
- `GET/POST /api/v1/companies` - List/Create companies
- `GET/PATCH/DELETE /api/v1/companies/:id` - Company CRUD
- `GET/POST /api/v1/deals` - List/Create deals
- `GET/PATCH/DELETE /api/v1/deals/:id` - Deal CRUD
- `GET /api/v1/deals/kanban` - Pipeline board view
- `GET/POST /api/v1/activities` - List/Create activities
- `GET /api/v1/timeline` - Entity timeline

## Project Structure

```
opencrm/
├── cmd/server/          # Entry point
├── internal/
│   ├── api/            # HTTP handlers and middleware
│   ├── domain/         # Business logic entities
│   ├── db/             # Migrations and queries
│   └── events/         # Event bus
├── pkg/                # Shared utilities
├── docker/             # Docker configuration
└── Makefile            # Development commands
```

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Generate code (sqlc)
make generate

# Create migration
make migrate/create name=add_new_table
```

## License

MIT
