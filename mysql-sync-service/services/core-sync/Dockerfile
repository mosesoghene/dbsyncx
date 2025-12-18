# Multi-stage build for smaller image
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy source code
COPY . .

# Generate go.sum and download dependencies
RUN go mod tidy

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sync-service ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/sync-service .

# Copy config (optional, usually mounted)
# COPY config.yaml .

EXPOSE 8080

CMD ["./sync-service"]
