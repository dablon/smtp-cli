# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git

# Copy source
COPY *.go ./
COPY go.mod ./

# Download dependencies (creates go.sum automatically)
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o smtp-cli -ldflags="-s -w" .

# Test with coverage
RUN go test -v -coverprofile=coverage.out -covermode=atomic .

# Show coverage
RUN go tool cover -func=coverage.out

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates mailx

WORKDIR /app

COPY --from=builder /build/smtp-cli /usr/local/bin/

RUN smtp-cli help

FROM builder AS test-only
CMD ["go", "test", "-v", "-coverprofile=coverage.out", "-covermode=atomic", "."]
