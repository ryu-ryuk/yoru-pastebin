.PHONY: setup start_db stop_db rm_db clean_db migrate_create migrate_up run build test clean

# Variables
DB_USER=ryu
DB_PASS=pass
DB_NAME=yoru_pastebin
DB_HOST=localhost
DB_PORT=5432
DB_CONTAINER_NAME=yoru-postgres
DB_CONNECTION_STRING=postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

MIGRATE_PATH=db/migrations
GO_CMD=go
MAIN_APP=./cmd/yoru/main.go

# Setup development environment (initial run)
setup: clean_db start_db migrate_up build

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
	# Run migrations
	# Ensure the migration files are present and correctly named (TIMESTAMP__NAME.up.sql / .down.sql)
	migrate -path $(MIGRATE_PATH) -database "$(DB_CONNECTION_STRING)" up
# Run the Go application
run:
	@echo "Running Yoru Pastebin..."
	$(GO_CMD) run $(MAIN_APP)

# Build the Go application binary
build:
	@echo "Building Yoru Pastebin binary..."
	$(GO_CMD) build -o yoru-pastebin $(MAIN_APP)

# Run tests (placeholder)
test:
	@echo "Running tests..."
	$(GO_CMD) test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f yoru-pastebin # Removes the built binary
	$(GO_CMD) clean