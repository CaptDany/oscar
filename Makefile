.PHONY: help build test lint migrate migrate-up migrate-down generate dev docker-up docker-down clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

## Build
build: ## Build the Go binary
	go build -o bin/oscar ./cmd/server

build/linux: ## Build for Linux (cross-compile)
	GOOS=linux GOARCH=amd64 go build -o bin/oscar-linux ./cmd/server

## Testing
test: ## Run unit tests
	go test -v -short ./...

test/integration: ## Run integration tests
	go test -v -tags=integration ./...

test/cover: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## Linting
lint: ## Run golangci-lint
	golangci-lint run ./...

lint/fix: ## Run golangci-lint with auto-fix
	golangci-lint run --fix ./...

## Code Generation
generate: ## Generate code (sqlc, etc)
	sqlc generate

generate/mocks: ## Generate mocks
	go generate ./...

## Database Migrations
migrate/up: ## Run pending migrations
	migrate -path internal/db/migrations -database "$$DATABASE_URL" up

migrate/down: ## Rollback last migration
	migrate -path internal/db/migrations -database "$$DATABASE_URL" down 1

migrate/create: ## Create new migration (usage: make migrate/create name=add_users_table)
	migrate create -ext sql -dir internal/db/migrations -seq $(name)

migrate/force: ## Force migration to version (usage: make migrate/force version=1)
	migrate -path internal/db/migrations -database "$$DATABASE_URL" force $(version)

## Development
dev: ## Run in development mode with hot reload
	air

run: ## Run the server
	go run ./cmd/server

## Docker
docker-up: ## Start all services with Docker Compose
	docker compose -f docker/docker-compose.yml up -d

docker-down: ## Stop all Docker Compose services
	docker compose -f docker/docker-compose.yml down

docker/build: ## Build Docker image
	docker build -f docker/Dockerfile -t opencrm:latest .

docker/push: ## Push Docker image to registry
	docker push opencrm:latest

## Cleanup
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

## Tools (install if missing)
tools: ## Install development tools
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/golang-migrate/migrate/cmd/migrate@latest
