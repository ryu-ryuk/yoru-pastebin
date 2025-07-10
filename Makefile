.PHONY: setup start_db stop_db rm_db clean_db migrate_create migrate_up run build test clean help prod-setup prod-deploy prod-status backup restore

# Variables - use environment variables if set, otherwise use defaults
DB_USER ?= $(if $(POSTGRES_USER),$(POSTGRES_USER),ryu)
DB_PASS ?= $(if $(POSTGRES_PASSWORD),$(POSTGRES_PASSWORD),pass)
DB_NAME ?= $(if $(POSTGRES_DB),$(POSTGRES_DB),yoru_pastebin)
DB_HOST ?= $(if $(POSTGRES_HOST),$(POSTGRES_HOST),localhost)
DB_PORT ?= $(if $(POSTGRES_PORT),$(POSTGRES_PORT),5432)
DB_CONTAINER_NAME=yoru-postgres
DB_CONNECTION_STRING=postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

MIGRATE_PATH=db/migrations
GO_CMD=go
MAIN_APP=./cmd/yoru/main.go

# Help target
help: ## Show this help message
	@echo "Yoru Pastebin - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

# Setup development environment (initial run)
setup: clean_db start_db migrate_up build ## Setup development environment

# starting the container and applying migrations
start_db:
	@echo "Starting PostgreSQL Docker container..."
	docker run --name $(DB_CONTAINER_NAME) \
	  -e POSTGRES_USER=$(DB_USER) \
	  -e POSTGRES_PASSWORD=$(DB_PASS) \
	  -e POSTGRES_DB=$(DB_NAME) \
	  -p $(DB_PORT):5432 \
	  -d postgres:16-alpine || echo "Container already running or exists."
	@echo "Waiting for PostgreSQL to be ready..."
	sleep 5 

# stop the postgres container
stop_db:
	@echo "Stopping PostgreSQL Docker container..."
	docker stop $(DB_CONTAINER_NAME) || true

# remove PostgreSQL 
rm_db:
	@echo "Removing PostgreSQL Docker container..."
	docker rm $(DB_CONTAINER_NAME) || true

# Clean database: Stop, remove container, then start fresh
clean_db: stop_db rm_db
	@echo "Database container cleaned."

# Usage: make migrate_create NAME=add_new_column
migrate_create:
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate_create NAME=<migration_name>"; exit 1; fi
	@echo "Creating new migration: $(NAME)"
	migrate create -ext sql -dir $(MIGRATE_PATH) $(NAME)

# Apply all pending migrations (called by `go run` but also runnable manually)
migrate_up:
	@echo "Applying database migrations..."
	# Ensure migrate CLI tool is available
	$(GO_CMD) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	# Run migrations using correct path
	$(HOME)/.local/share/go/bin/migrate -path $(MIGRATE_PATH) -database "$(DB_CONNECTION_STRING)" up
	
# Run the Go application for local development
run:
	@echo "Running Yoru Pastebin (local development)..."
	@echo "Using local database configuration..."
	DATABASE_CONNECTION_STRING="$(DB_CONNECTION_STRING)" \
	SERVER_PORT=8000 \
	BASE_URL="http://localhost:8081" \
	$(GO_CMD) run $(MAIN_APP)

# Run the Go application using production environment variables
run-prod:
	@echo "Running Yoru Pastebin (production mode)..."
	@echo "Using .env file for configuration..."
	@if [ -f .env ]; then \
		export $$(grep -v '^#' .env | xargs) && \
		DATABASE_CONNECTION_STRING="postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" \
		$(GO_CMD) run $(MAIN_APP); \
	else \
		echo "Error: .env file not found!"; \
		exit 1; \
	fi

# Build the Go application binary
build:
	@echo "Building Yoru Pastebin binary..."
	$(GO_CMD) build -o yoru-pastebin $(MAIN_APP)

# Run tests (placeholder)
test: ## Run tests
	@echo "Running tests..."
	$(GO_CMD) test -v ./...

# Production commands
prod-setup: ## Setup production environment
	@echo "Setting up production environment..."
	@./deploy.sh setup

prod-deploy: ## Deploy to production
	@echo "Deploying to production..."
	@./deploy.sh deploy

prod-status: ## Show production status
	@./deploy.sh status

prod-logs: ## Show production logs
	@./deploy.sh logs

prod-restart: ## Restart production services
	@./deploy.sh restart

prod-stop: ## Stop production services
	@./deploy.sh stop

# Database operations
backup: ## Create database backup
	@echo "Creating database backup..."
	@./deploy.sh backup

restore: ## Restore database from backup (requires BACKUP_FILE=path)
	@echo "Restoring database..."
	@./deploy.sh restore $(BACKUP_FILE)

# Monitoring
health: ## Check service health
	@./deploy.sh health

# Security
security-scan: ## Run security scan on Docker images
	@echo "Running security scan..."
	@docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
		aquasec/trivy image yoru-pastebin:latest

# Quick commands
quick-deploy: build prod-deploy ## Quick build and deploy