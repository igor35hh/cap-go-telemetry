# Makefile for cap-go-telemetry

# Variables
GO_VERSION := 1.23
PROJECT_NAME := cap-go-telemetry
PACKAGE := github.com/iklimetscisco/cap-go-telemetry
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -ldflags "-X $(PACKAGE)/internal/version.Version=$(VERSION) \
                     -X $(PACKAGE)/internal/version.GitCommit=$(GIT_COMMIT) \
                     -X $(PACKAGE)/internal/version.BuildDate=$(BUILD_DATE)"

# Directories
BUILD_DIR := build
EXAMPLES_DIR := examples

.PHONY: all build test clean lint fmt vet deps examples help

# Default target
all: fmt vet test build

# Build all examples
build: deps
	@echo "Building examples..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/basic $(EXAMPLES_DIR)/basic/main.go
	@echo "Build complete. Binaries are in $(BUILD_DIR)/"

# Run tests
test: deps
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean -cache -testcache -modcache

# Lint code
lint: deps
	@echo "Running linters..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	}
	@golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	@go vet ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Build and run examples
examples: build
	@echo "Building examples..."
	@echo "To run the basic example:"
	@echo "  ./$(BUILD_DIR)/basic"

# Run the basic example
run-basic: build
	@echo "Running basic example..."
	@./$(BUILD_DIR)/basic

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "Development setup complete!"

# Generate documentation
docs:
	@echo "Generating documentation..."
	@command -v godoc >/dev/null 2>&1 || go install golang.org/x/tools/cmd/godoc@latest
	@echo "Run 'godoc -http=:6060' and visit http://localhost:6060/pkg/$(PACKAGE)/"

# Security scan
security:
	@echo "Running security scan..."
	@command -v gosec >/dev/null 2>&1 || go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@gosec ./...

# Check for vulnerabilities
vuln-check:
	@echo "Checking for vulnerabilities..."
	@command -v govulncheck >/dev/null 2>&1 || go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Release preparation
release-check: fmt vet lint test security vuln-check
	@echo "Release checks passed!"

# Show version information
version:
	@echo "Project: $(PROJECT_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Go Version: $(GO_VERSION)"

# Help
help:
	@echo "Available targets:"
	@echo "  all           - Run fmt, vet, test, and build"
	@echo "  build         - Build all examples"
	@echo "  test          - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  lint          - Run linters"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  deps          - Install dependencies"
	@echo "  examples      - Build examples"
	@echo "  run-basic     - Build and run basic example"
	@echo "  dev-setup     - Set up development environment"
	@echo "  docs          - Generate documentation"
	@echo "  security      - Run security scan"
	@echo "  vuln-check    - Check for vulnerabilities"
	@echo "  update-deps   - Update dependencies"
	@echo "  release-check - Run all checks for release"
	@echo "  version       - Show version information"
	@echo "  help          - Show this help message"
