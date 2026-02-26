# SMTP CLI

A command-line tool for sending emails via SMTP servers.

## Features

- Send plain text and HTML emails
- SMTP authentication support
- Connection testing
- Configurable via flags or environment variables
- Docker support for easy deployment

## Installation

### From Source

```bash
go build -o smtp-cli .
```

### Using Docker

```bash
docker build -t smtp-cli .
```

## Usage

### Send an Email

```bash
# Plain text email
smtp-cli send --to user@example.com --subject "Hello" --body "Message body"

# HTML email
smtp-cli send --to user@example.com --subject "Hello" --html "<h1>Hello!</h1>"

# With authentication
smtp-cli send --host smtp.maleon.run --port 5870 --user elus54 --pass yourpass --to user@example.com --subject "Test" --body "Message"
```

### Test Connection

```bash
smtp-cli test --host smtp.maleon.run --port 5870
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| SMTP_HOST | SMTP server hostname | smtp.maleon.run |
| SMTP_PORT | SMTP server port | 5870 |
| SMTP_USER | SMTP username | elus54 |
| SMTP_PASS | SMTP password | (none) |

## Docker Compose

```bash
# Start SMTP server and CLI
docker-compose up -d

# Run CLI
docker exec smtp-cli smtp-cli send --to user@example.com --subject "Test" --body "Hello"
```

## Testing

```bash
# Run unit tests
go test -v -coverprofile=coverage.out ./...

# Run E2E tests (requires SMTP server)
SMTP_HOST=smtp.maleon.run SMTP_PORT=5870 SMTP_USER=elus54 SMTP_PASS=pass go test -v -tags=e2e ./...
```

## Coverage

Current coverage: >90%

## License

MIT
