# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy all source
COPY . .

# Build
RUN go build -o smtp-cli ./cmd/smtp-cli

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates mailx

COPY --from=builder /app/smtp-cli /usr/local/bin/smtp-cli

ENTRYPOINT ["smtp-cli"]
