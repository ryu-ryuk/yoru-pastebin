# ----------------------------
# Stage 1: Build the Go binary
# ----------------------------
FROM golang:1.24.4-alpine AS builder

# Install necessary tools
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files and download deps first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build optimized binary
# -trimpath: removes local path info
# -ldflags="-s -w": strips debug info for smaller binary
# CGO_ENABLED=0: static binary (no C dependencies)
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /yoru-pastebin ./cmd/yoru/main.go

# ----------------------------
# Stage 2: Final minimal image
# ----------------------------
FROM alpine:3.20

# Install root CA certs for HTTPS support (needed for PostgreSQL)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy built Go binary
COPY --from=builder /yoru-pastebin .

# Copy config, templates, static files, migrations
COPY configs/ config/
COPY web/templates/ web/templates/
COPY web/static/ web/static/
COPY db/migrations/ db/migrations/

# Expose the port configured in your app (.env or config.toml)
EXPOSE 8080

# Run the binary
CMD ["./yoru-pastebin"]
