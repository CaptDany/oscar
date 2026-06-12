# Contributing to Oscar

## Documentation

See the [Documentation Index](docs/README.md) for the full set of project docs,
including architecture, API reference, CI/CD pipeline, and project management artifacts.

## Development Setup

### Prerequisites
- Go 1.24+
- Node.js 22+
- PostgreSQL 16+
- Redis

### Local Setup
```bash
# Clone the repository
git clone https://github.com/CaptDany/oscar.git
cd oscar

# Copy and configure environment
cp .env.example .env

# Start dependencies (PostgreSQL, Redis)
docker compose up -d

# Run database migrations
go run cmd/migrate/main.go up

# Start the backend
go run cmd/server/main.go

# In a separate terminal, start the frontend
cd web
npm install
npm run dev
```

### Running Tests
```bash
# Backend tests
go test -v -short -count=1 ./...

# Frontend build check
cd web
npm run build
```

### Linting
```bash
# Backend
golangci-lint run ./...

# Frontend
cd web
npm run build  # includes type checking
```

## Code Style

### Go
- Follow standard Go conventions (`gofmt`, `golint`)
- Use `errs` package for consistent error responses
- Repository pattern: domain models → repository interfaces → sqlc/pgx implementations
- Handler methods should be thin (parse request → call service → marshal response)

### Frontend (Astro + React/Preact)
- Use Preact with compat for React-compatible components
- Prefer inline `<script>` with `fetch()` for API calls over heavy state management
- Follow existing patterns in settings pages for new admin UI
- TypeScript for all new code

## Commit Conventions
We follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` — New feature
- `fix:` — Bug fix
- `chore:` — Maintenance, dependencies, tooling
- `docs:` — Documentation only
- `refactor:` — Code restructuring without feature/bug changes
- `test:` — Adding or fixing tests

## Pull Request Process
1. Create a branch from `main` with a descriptive name:
   - `feat/my-feature` for new features
   - `fix/my-bugfix` for bug fixes
   - `chore/my-task` for maintenance
2. Make your changes, keeping commits small and focused
3. Ensure all tests pass and lint is clean
4. Open a PR against `main` with a clear description
5. Request review from a maintainer
6. Address any feedback and keep the branch up to date

## Code of Conduct
Please note that this project follows a [Code of Conduct](CODE_OF_CONDUCT.md).
By participating, you agree to uphold its standards.
