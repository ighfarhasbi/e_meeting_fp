# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN go build -o main ./cmd/server

# Stage 2: Run
FROM alpine:3.20

WORKDIR /app

# Copy only binary from builder
COPY --from=builder /app/main .

# Expose port REST API
EXPOSE 8080

# Jalankan aplikasi
CMD ["./main"]
