# Stage 1: Build the Go application
FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
# CGO_ENABLED=0 is important for static binaries
# -a -installsuffix cgo reduces binary size
# -ldflags="-s -w" removes debugging information for smaller binary
RUN CGO_ENABLED=0 go build -o /yoru-pastebin ./cmd/yoru/main.go

# Stage 2: Create the final lean image
FROM alpine:latest

# Install ca-certificates for HTTPS connections (needed by pgx)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /yoru-pastebin .

# Copy configuration and static/template files
COPY configs/ config/
COPY web/templates/ web/templates/
COPY web/static/ web/static/

# Expose the port your server listens on (from config.toml:8080)
EXPOSE 8080

# Command to run the application
# Use -c /root/config/config.toml if you hardcode the config file path,
# or ensure your app looks in the current working directory for 'configs/config.toml'
# Our LoadConfig in Go searches "./configs", so this is fine.
CMD ["./yoru-pastebin"]