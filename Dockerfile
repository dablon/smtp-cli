# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy all source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o smtp-cli -ldflags="-s -w" ./cmd/smtp-cli

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates mailx

WORKDIR /app

COPY --from=builder /build/smtp-cli /usr/local/bin/

CMD ["smtp-cli", "help"]
