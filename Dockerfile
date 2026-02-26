# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go mod files first
COPY go.mod ./
RUN go mod download

# Copy all source
COPY . .

# Build - the module is at root, so build from there
RUN CGO_ENABLED=0 GOOS=linux go build -o smtp-cli ./cmd/smtp-cli

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates mailx

WORKDIR /app

COPY --from=builder /build/smtp-cli /usr/local/bin/

CMD ["smtp-cli", "help"]
