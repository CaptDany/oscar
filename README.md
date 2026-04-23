# oscar CRM

A production-grade, open-source CRM backend built in Go with multi-tenant SaaS architecture, designed for scalability and performance.

## Features

### Core CRM
- **Multi-tenant Architecture**: Complete tenant isolation with Row Level Security (RLS)
- **Contacts Management**: Persons (leads, contacts, customers) with tags, scores, and custom fields
- **Company Management**: Track companies with industry, size, revenue, and associations
- **Deal Pipeline**: Kanban boards, multiple pipelines, stages, and probability tracking
- **Activity Tracking**: Notes, calls, emails, meetings, tasks with timeline view
- **WebSocket Support**: Real-time updates for live collaboration

### Security & Auth
- **Paseto v2 Authentication**: Stateless, secure token-based authentication
- **Role-Based Access Control**: Flexible permission system (Owner, Admin, Manager, Sales Rep, Read Only)
- **Row Level Security**: Database-level tenant isolation for maximum security
- **Multi-factor Ready**: Architecture supports future MFA implementation

### Customization
- **Custom Fields**: Define custom fields for any entity (persons, companies, deals, activities)
- **White-label Branding**: Custom logos, colors, fonts per tenant
- **Automation Engine**: Trigger-based workflows with parallel action execution
- **Webhook Support**: Integrate with external systems

### Developer Experience
- **sqlc Integration**: Type-safe SQL queries with code generation
- **Repository Pattern**: Clean separation between data access and business logic
- **Comprehensive Error Handling**: Structured error responses with HTTPError interface
- **Cursor-based Pagination**: Efficient pagination for large datasets
- **OpenTelemetry**: Distributed tracing support

## Tech Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Language | Go 1.23+ | Backend API |
| Framework | Echo v4 | HTTP routing |
| Database | PostgreSQL 16 | Primary data store |
| ORM/Driver | pgx/v5 | Database access |
| Code Gen | sqlc v1.30.0 | Type-safe SQL |
| Auth | Paseto v2 | Token authentication |
| Validation | go-playground/validator | Request validation |
| Testing | testify | Unit testing |

## Project Structure

```
oscar/
├── cmd/
│   └── server/              # Application entry point
│       └── main.go           # Server initialization and wiring
│
├── internal/
│   ├── api/                 # HTTP layer
│   │   ├── handlers/        # Request handlers
│   │   │   ├── auth.go      # Authentication endpoints
│   │   │   ├── persons.go   # Person CRUD operations
│   │   │   ├── companies.go # Company CRUD operations
│   │   │   ├── deals.go     # Deal and pipeline operations
│   │   │   ├── pipelines.go # Pipeline management
│   │   │   └── activities.go# Activity tracking
│   │   ├── middleware/      # HTTP middleware
│   │   │   └── middleware.go# Auth, tenant resolution, rate limiting
│   │   ├── routes.go        # Route definitions
│   │   ├── server.go        # Echo server setup
│   │   └── ws/              # WebSocket support
│   │       └── handler.go   # Real-time communication
│   │
│   ├── config/              # Configuration management
│   │   └── config.go        # Env var loading
│   │
│   ├── db/                  # Database layer
│   │   ├── generated/       # sqlc generated code
│   │   │   ├── models.go    # Generated model types
│   │   │   └── *.sql.go    # Generated query functions
│   │   ├── repositories/    # Data access layer
│   │   │   ├── person.go    # Person repository
│   │   │   ├── company.go   # Company repository
│   │   │   ├── deal.go      # Deal & pipeline repository
│   │   │   ├── activity.go  # Activity & association repository
│   │   │   ├── team.go      # Team management
│   │   │   ├── tenant.go    # Tenant & branding
│   │   │   ├── user.go      # User & role management
│   │   │   ├── custom_field.go
│   │   │   ├── automation.go # Automation rules
│   │   │   ├── notification.go
│   │   │   ├── audit_log.go
│   │   │   └── helpers.go   # Type conversion utilities
│   │   └── schema.sql       # Database schema
│   │
│   ├── domain/              # Business logic layer
│   │   ├── person/
│   │   │   └── person.go   # Person types & interfaces
│   │   ├── company/
│   │   │   └── company.go   # Company types & interfaces
│   │   ├── deal/
│   │   │   └── deal.go      # Deal, Pipeline types
│   │   ├── activity/
│   │   │   └── activity.go  # Activity types
│   │   ├── team/
│   │   │   └── team.go      # Team types
│   │   ├── tenant/
│   │   │   └── tenant.go    # Tenant & branding types
│   │   ├── user/
│   │   │   └── user.go      # User, Role, Permission types
│   │   ├── custom_field/
│   │   │   └── custom_field.go
│   │   ├── automation/
│   │   │   └── automation.go # Automation types
│   │   ├── notification/
│   │   │   └── notification.go
│   │   ├── audit_log/
│   │   │   └── audit_log.go
│   │   └── product/
│   │       └── product.go   # Product types
│   │
│   ├── events/              # Event bus
│   │   └── events.go        # Event definitions
│   │
│   ├── email/               # Email service
│   │   └── email.go
│   │
│   └── storage/             # File storage
│       └── storage.go
│
├── pkg/                     # Shared packages
│   ├── crypto/
│   │   ├── crypto.go       # Password hashing, API keys
│   │   └── token.go        # Paseto token management
│   ├── errs/
│   │   └── errors.go       # Structured errors
│   ├── validator/
│   │   └── validator.go     # Custom validators
│   └── pagination/
│       └── pagination.go    # Pagination utilities
│
├── docker/                  # Docker configuration
│   ├── Dockerfile
│   └── docker-compose.yml
│
├── Makefile                # Development commands
├── sqlc.yaml               # sqlc configuration
└── go.mod                  # Go dependencies
```

## Architecture

### Layered Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Layer (Echo)                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐│
│  │ Middleware  │→ │  Handlers   │→ │   Request/Response   ││
│  └─────────────┘  └─────────────┘  └─────────────────────┘│
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   Business Logic Layer                        │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              Domain Types & Interfaces                   ││
│  │  (Person, Company, Deal, Activity, Tenant, User, etc.) ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                     Data Access Layer                        │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                    Repositories                          ││
│  │  (PersonRepo, CompanyRepo, DealRepo, etc.)             ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      Database Layer                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐│
│  │ PostgreSQL  │  │   pgx/v5   │  │  sqlc Generated     ││
│  └─────────────┘  └─────────────┘  └─────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

### Repository Pattern

Each domain entity follows the repository pattern with a clear interface:

```go
// Domain defines the repository interface
type Repository interface {
    Create(ctx context.Context, tenantID uuid.UUID, req *CreateRequest) (*Entity, error)
    GetByID(ctx context.Context, id uuid.UUID) (*Entity, error)
    List(ctx context.Context, tenantID uuid.UUID, filter *Filter) ([]*Entity, string, int, error)
    Update(ctx context.Context, id uuid.UUID, req *UpdateRequest) (*Entity, error)
    Delete(ctx context.Context, id uuid.UUID) error
}

// Repository implementation in data layer
type EntityRepository struct {
    pool *pgxpool.Pool
}
```

### Error Wrapping Convention

All errors follow a consistent wrapping format for traceability:

```go
fmt.Errorf("domain.Method: %w", err)
```

Examples:
- `fmt.Errorf("person.Create: %w", err)`
- `fmt.Errorf("deal.Update: %w", err)`
- `fmt.Errorf("user.GetByEmail: %w", err)`

## Database Schema

### Key Tables

- `tenants` - Multi-tenant support
- `users` - User accounts with password hashing
- `roles` - Role definitions with permissions
- `persons` - Leads, contacts, customers
- `companies` - Company records
- `deals` - Sales opportunities
- `pipelines` - Deal pipelines
- `pipeline_stages` - Pipeline stages
- `activities` - Activity log
- `activity_associations` - Activity-entity links
- `custom_field_definitions` - Custom field schemas
- `automations` - Automation rules
- `automation_actions` - Automation action steps
- `automation_runs` - Automation execution logs
- `notifications` - User notifications
- `audit_logs` - Audit trail
- `team_members` - Team memberships

### Row Level Security

PostgreSQL RLS ensures complete tenant isolation:

```sql
CREATE POLICY tenant_isolation ON persons
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new tenant with first user |
| POST | `/api/v1/auth/login` | Authenticate and receive tokens |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Invalidate session |
| GET | `/api/v1/auth/me` | Get current user info |

### Persons (CRM)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/persons` | List persons with filtering |
| POST | `/api/v1/persons` | Create new person |
| GET | `/api/v1/persons/:id` | Get person by ID |
| PATCH | `/api/v1/persons/:id` | Update person |
| DELETE | `/api/v1/persons/:id` | Soft delete person |
| POST | `/api/v1/persons/:id/convert` | Convert lead to contact |
| POST | `/api/v1/persons/:id/tags` | Add tag to person |
| DELETE | `/api/v1/persons/:id/tags` | Remove tag |
| GET | `/api/v1/persons/search` | Search persons |

### Companies

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/companies` | List companies |
| POST | `/api/v1/companies` | Create company |
| GET | `/api/v1/companies/:id` | Get company |
| PATCH | `/api/v1/companies/:id` | Update company |
| DELETE | `/api/v1/companies/:id` | Delete company |

### Deals & Pipelines

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/deals` | List deals |
| POST | `/api/v1/deals` | Create deal |
| GET | `/api/v1/deals/kanban` | Kanban board view |
| GET | `/api/v1/deals/:id` | Get deal |
| PATCH | `/api/v1/deals/:id` | Update deal |
| DELETE | `/api/v1/deals/:id` | Delete deal |
| PATCH | `/api/v1/deals/:id/stage` | Move to stage |
| POST | `/api/v1/deals/:id/win` | Close as won |
| POST | `/api/v1/deals/:id/lose` | Close as lost |

### Pipelines

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/pipelines` | List pipelines |
| POST | `/api/v1/pipelines` | Create pipeline |
| GET | `/api/v1/pipelines/:id` | Get pipeline |
| PATCH | `/api/v1/pipelines/:id` | Update pipeline |
| DELETE | `/api/v1/pipelines/:id` | Delete pipeline |
| GET | `/api/v1/pipelines/:id/stages` | List stages |
| POST | `/api/v1/pipelines/:id/stages` | Create stage |
| PATCH | `/api/v1/pipelines/:id/stages/reorder` | Reorder stages |
| PATCH | `/api/v1/pipelines/:id/stages/:stage_id` | Update stage |
| DELETE | `/api/v1/pipelines/:id/stages/:stage_id` | Delete stage |

### Activities

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/activities` | List activities |
| POST | `/api/v1/activities` | Create activity |
| GET | `/api/v1/activities/:id` | Get activity |
| PATCH | `/api/v1/activities/:id` | Update activity |
| POST | `/api/v1/activities/:id/complete` | Mark complete |
| DELETE | `/api/v1/activities/:id` | Delete activity |
| GET | `/api/v1/timeline` | Entity timeline |

## Quick Start

### Prerequisites

- Go 1.23 or later
- PostgreSQL 16+
- Make

### Installation

```bash
# Clone the repository
git clone https://github.com/oscar/oscar.git
cd oscar

# Install dependencies
go mod download

# Copy environment file
cp .env.example .env

# Edit .env with your database credentials
```

### Database Setup

```bash
# Apply migrations
make migrate/up

# Seed initial data (optional)
make seed
```

### Run Development Server

```bash
# Start the server
go run ./cmd/server

# Or use make
make dev
```

The server starts on `http://localhost:8080`

### Run Tests

```bash
make test

# Or directly
go test ./...
```

## Configuration

Environment variables (see `.env.example`):

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_SECRET` | JWT signing secret | - |
| `APP_HOST` | Server host | `0.0.0.0` |
| `APP_PORT` | Server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | - |

## Authentication Flow

1. **Register**: `POST /auth/register` creates tenant + user
2. **Login**: `POST /auth/login` returns access + refresh tokens
3. **Authenticate**: Include `Authorization: Bearer <token>` header
4. **Refresh**: `POST /auth/refresh` with refresh token

## Development

### Generate SQL Code

```bash
# Generate repository code from SQL
make generate

# Watch mode for development
make generate-watch
```

### Create Migration

```bash
make migrate/create name=add_new_column
```

### Code Quality

```bash
# Run linter
make lint

# Format code
make fmt

# Vet code
go vet ./...
```

## Testing

Tests use the `testify` framework:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/domain/person/...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## License

GNU GPLv3 - see LICENSE file for details.

## Roadmap

- [ ] Redis integration for caching and sessions
- [ ] Email/SMS notification delivery
- [ ] Import/Export CSV functionality
- [ ] Advanced automation conditions
- [ ] Audit log API
- [ ] Dashboard analytics
- [ ] Mobile push notifications
