.PHONY: help dev lint test build run clean docker-build docker-run

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Run the application in development mode with live reload
	@echo "Starting development server..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	@air

lint: ## Run linter
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

build: ## Build the application
	@echo "Building application..."
	@go build -o bin/leetbot ./cmd/bot

run: build ## Build and run the application
	@echo "Running application..."
	@./bin/leetbot

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t leetbot .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run --env-file .env leetbot

validate: lint test ## Run linting and tests
	@echo "Validation complete!"

setup: ## Setup development environment
	@echo "Setting up development environment..."
	@go mod tidy
	@go mod download
	@echo "Installing development tools..."
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Setup complete!"

generate-embedded: ## Generate embedded CSV data from actual files
	@echo "Generating embedded data..."
	@go run scripts/generate_embedded/main.go data
	@echo "Copy the generated content to internal/data/parser.go"

validate-data: ## Validate all CSV files in data directory
	@echo "Validating CSV data..."
	@go run scripts/validate_data/main.go data

demo: ## Run the bot demo
	@echo "Running bot demo..."
	@go run scripts/demo_bot/main.go
