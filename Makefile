.PHONY: all build run test lint clean docker-build docker-up docker-down help

APP_NAME := identity-validation-mx
VERSION := 1.0.0
DOCKER_IMAGE := $(APP_NAME):$(VERSION)
GO := go
GOFLAGS := -v

all: build

## Build the application
build:
	@echo "Building $(APP_NAME)..."
	$(GO) build $(GOFLAGS) -o bin/$(APP_NAME) ./cmd/server

## Run the application
run:
	@echo "Running $(APP_NAME)..."
	$(GO) run ./cmd/server

## Run tests
test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) ./...

## Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test $(GOFLAGS) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

## Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

## Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	$(GO) clean

## Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

## Build Docker image
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) -f deployments/docker/Dockerfile .

## Run Docker containers
docker-up:
	@echo "Starting Docker containers..."
	docker-compose -f deployments/docker/docker-compose.yml up -d

## Stop Docker containers
docker-down:
	@echo "Stopping Docker containers..."
	docker-compose -f deployments/docker/docker-compose.yml down

## Run database migrations
migrate-up:
	@echo "Running migrations..."
	$(GO) run ./cmd/migrate up

## Rollback database migrations
migrate-down:
	@echo "Rollback migrations..."
	$(GO) run ./cmd/migrate down

## Generate Swagger docs
swagger:
	@echo "Generating Swagger docs..."
	swag init -g cmd/server/main.go -o ./docs

## Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## Show help
help:
	@echo "Available targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'

.DEFAULT_GOAL := help
