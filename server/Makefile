.PHONY: test test-unit test-integration build clean

# Build the server binary
build:
	go build -o server ./cmd/server

# Run all tests (unit + integration)
test: test-unit test-integration

# Run unit tests only (skip integration tests)
test-unit:
	go test -v -short ./...

# Run integration tests with testcontainers
# Requires Docker to be running
test-integration:
	go test -v -tags=integration ./integration/...

# Run integration tests with verbose output and no caching
test-integration-nocache:
	go test -v -count=1 -tags=integration ./integration/...

# Run a specific integration test
# Usage: make test-integration-run TEST=TestTeamRepository_CreateAndGetByID
test-integration-run:
	go test -v -tags=integration -run $(TEST) ./integration/...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run integration tests with coverage
test-integration-coverage:
	go test -v -tags=integration -coverprofile=integration-coverage.out ./integration/...
	go tool cover -html=integration-coverage.out -o integration-coverage.html

# Clean build artifacts
clean:
	rm -f server coverage.out coverage.html integration-coverage.out integration-coverage.html

# Tidy dependencies
tidy:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Run all checks (format, lint, test)
check: fmt lint test
