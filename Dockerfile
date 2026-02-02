# Stage 1: Build the application
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git gcc

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project source
COPY . .

# Build the application
# Disabling CGO for a fully static binary (easier for alpine/scratch)
# Pointing to the main entry point in cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/gateway ./cmd/main.go

# Stage 2: Final runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/gateway .
# Copy .env or other assets if required at runtime
COPY --from=builder /app/.env .

EXPOSE 8080

# Run the binary
ENTRYPOINT ["./gateway"]
CMD ["./cmd/main"]