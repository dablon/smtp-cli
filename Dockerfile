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

# Build only (skip tests for faster builds)
RUN CGO_ENABLED=0 GOOS=linux go build -o smtp-cli -ldflags="-s -w" .

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates mailx

WORKDIR /app

COPY --from=builder /build/smtp-cli /usr/local/bin/

CMD ["smtp-cli", "help"]
