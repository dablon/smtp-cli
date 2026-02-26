# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy all source at once
COPY . .

# Build
RUN go build -o smtp-cli ./cmd/smtp-cli

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates mailx

WORKDIR /app

COPY --from=builder /app/smtp-cli /usr/local/bin/

CMD ["smtp-cli", "help"]
