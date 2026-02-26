# SMTP CLI

A command-line tool for sending emails via SMTP servers, written in Go.

[![CI](https://github.com/dablon/smtp-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/dablon/smtp-cli/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/dablon/smtp-cli)](https://github.com/dablon/smtp-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- Send plain text and HTML emails
- SMTP authentication support (PLAIN/LOGIN)
- TLS/STARTTLS support
- Connection testing
- Configurable via flags or environment variables
- Docker support for easy deployment

## Installation

### From Source

```bash
git clone https://github.com/dablon/smtp-cli.git
cd smtp-cli
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
smtp-cli send -to user@example.com -subject "Hello" -body "Message body"

# HTML email
smtp-cli send -to user@example.com -subject "Hello" -html "<h1>Hello!</h1>"

# With SMTP authentication
smtp-cli send -host smtp.example.com -port 587 -user myuser -pass mypass -to user@example.com -subject "Test" -body "Message"
```

### Test Connection

```bash
smtp-cli test -host smtp.example.com -port 587
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-host` | SMTP server hostname | smtp.maleon.run |
| `-port` | SMTP server port | 5870 |
| `-user` | SMTP username | (none) |
| `-pass` | SMTP password | (none) |
| `-from` | From address | noreply@maleon.run |
| `-to` | To address | (required) |
| `-subject` | Email subject | (none) |
| `-body` | Plain text body | (none) |
| `-html` | HTML body | (none) |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `SMTP_HOST` | SMTP server hostname |
| `SMTP_PORT` | SMTP server port |
| `SMTP_USER` | SMTP username |
| `SMTP_PASS` | SMTP password |

## Docker Compose

```bash
# Start SMTP server and CLI
docker-compose up -d

# Run CLI
docker exec smtp-cli smtp-cli send -to user@example.com -subject "Test" -body "Hello"
```

## Testing

```bash
# Run unit tests
go test -v ./...

# Run with coverage
go test -v -coverprofile=coverage.out -covermode=atomic ./...

# Run E2E tests (requires SMTP server)
SMTP_HOST=localhost SMTP_PORT=2525 go test -v -tags=e2e ./...
```

## License

MIT License - see [LICENSE](LICENSE) for details.
