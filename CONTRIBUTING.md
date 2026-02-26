# Contributing to SMTP CLI

Thank you for your interest in contributing!

## Development Setup

### Prerequisites
- Go 1.21+
- Docker (optional, for containerized testing)

### Local Development

```bash
# Clone the repo
git clone https://github.com/dablon/smtp-cli.git
cd smtp-cli

# Build
go build -o smtp-cli .

# Run tests
go test -v ./...

# Run with coverage
go test -v -coverprofile=coverage.out -covermode=atomic ./...
```

### Docker Development

```bash
# Build image
docker build -t smtp-cli .

# Run tests in container
docker run --rm smtp-cli smtp-cli help

# Run with Docker Compose
docker-compose up -d
```

## Code Style

- Use `go fmt` before committing
- Add tests for new features
- Keep coverage above 80%

## Testing

```bash
# Unit tests
go test -v ./...

# E2E tests (requires SMTP server)
SMTP_HOST=localhost SMTP_PORT=2525 go test -v -tags=e2e ./...
```

## Submitting Changes

1. Create a feature branch
2. Make your changes
3. Add tests
4. Ensure CI passes
5. Submit a pull request
