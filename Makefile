.PHONY: build test test-unit test-e2e docker-build docker-run clean coverage

# Build the CLI
build:
	go build -o smtp-cli ./cmd/smtp-cli

# Run all tests
test: test-unit

# Run unit tests with coverage
test-unit:
	go test -v -coverprofile=coverage.out -covermode=atomic ./...

# Run E2E tests (requires SMTP server running)
test-e2e:
	go test -v -tags=e2e -run E2E ./...

# Run tests in Docker
docker-test:
	docker build -t smtp-cli:test --target test-only .

# Build Docker image
docker-build:
	docker build -t smtp-cli .

# Run CLI in Docker
docker-run:
	docker run --rm smtp-cli help

# Show coverage
coverage:
	go tool cover -func=coverage.out

# Clean build artifacts
clean:
	rm -f smtp-cli coverage.out

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o smtp-cli-linux ./cmd/smtp-cli
	GOOS=darwin GOARCH=amd64 go build -o smtp-cli-macos ./cmd/smtp-cli
	GOOS=windows GOARCH=amd64 go build -o smtp-cli.exe ./cmd/smtp-cli
