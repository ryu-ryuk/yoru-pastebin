
<h1 align="center">
  <img src="https://raw.githubusercontent.com/ryu-ryuk/yoru-pastebin/main/docs/assets/yoru_logo.png" width="800" alt="Yoru Pastebin Banner"/>
  <img src="https://raw.githubusercontent.com/catppuccin/catppuccin/main/assets/misc/transparent.png" height="16" width="0px"/>
  
  <span style="color:#cdd6f4;">Yoru Pastebin</span>
</h1>

<h6 align="center" style="color:#bac2de;">
  A production-grade, security-first pastebin with enterprise features
</h6>

<p align="center">
  <a href="https://github.com/ryu-ryuk/yoru-pastebin/stargazers"><img src="https://img.shields.io/github/stars/ryu-ryuk/yoru-pastebin?colorA=1e1e2e&colorB=cba6f7&style=for-the-badge&logo=github&logoColor=cdd6f4"></a><a href="https://github.com/ryu-ryuk/yoru-pastebin/issues"><img src="https://img.shields.io/github/issues/ryu-ryuk/yoru-pastebin?colorA=1e1e2e&colorB=f38ba8&style=for-the-badge&logo=github&logoColor=cdd6f4"></a><a href="https://github.com/ryu-ryuk/yoru-pastebin/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-GPLv3-89b4fa?style=for-the-badge&logo=gnu&logoColor=1e1e2e&colorA=1e1e2e"></a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-89b4fa?style=for-the-badge&logo=go&logoColor=white&colorA=1e1e2e" />
  <img src="https://img.shields.io/badge/PostgreSQL-15+-b4befe?style=for-the-badge&logo=postgresql&logoColor=white&colorA=1e1e2e" />
  <img src="https://img.shields.io/badge/Docker-Production-94e2d5?style=for-the-badge&logo=docker&logoColor=white&colorA=1e1e2e" />
  <img src="https://img.shields.io/badge/Traefik-SSL-fab387?style=for-the-badge&logo=traefikmesh&logoColor=white&colorA=1e1e2e" />
  <img src="https://img.shields.io/badge/File%20Storage-Hybrid-94e2d5?style=for-the-badge&logo=files&logoColor=white&colorA=1e1e2e" />
</p>


**A production-grade, security-hardened pastebin service built for teams and organizations.**

Yoru Pastebin implements enterprise-grade security, performance optimizations, and production deployment capabilities. Built with Go, PostgreSQL, deployed on GCP with support for hybrid cloud storage.

**Live Demo:** [https://paste.alokranjan.me](https://paste.alokranjan.me)

## Key Features

### Security Architecture
- **Zero-knowledge encryption** with AES-256-GCM for password-protected pastes
- **Cryptographically secure ID generation** using Base62 encoding
- **PBKDF2 key derivation** with bcrypt cost factor 12 for password hashing
- **Content Security Policy** headers and XSS protection via Traefik
- **Rate limiting** (2 requests/second default) and DDoS protection
- **Automatic expiration** with secure deletion (24-hour default)

### Performance Optimizations
- **Lazy-loaded syntax highlighting** with 35+ language support
- **Optimized Go backend** with efficient request handling
- **PostgreSQL 16** with connection pooling
- **Multi-stage Docker builds** for minimal image size
- **Static asset optimization** with proper caching headers

### Production Infrastructure
- **Multi-replica deployment** with Docker Swarm/Compose
- **Hybrid file storage** using local volumes and optional AWS S3 (20MB max per file)
- **PostgreSQL persistence** with volume mounting
- **Graceful shutdown** handling with 10-second timeout
- **Health monitoring** and container orchestration

### Developer Experience
- **RESTful API** for programmatic access
- **File upload support** with drag-and-drop interface
- **35+ programming languages** with syntax highlighting
- **Responsive web interface** with modern catppuccin UI
- **Comprehensive Makefile** for development workflow
- **Database migrations** with version control

## Quick Start

### Development Setup
```bash
# Clone repository
git clone https://github.com/ryu-ryuk/yoru-pastebin.git
cd yoru-pastebin

# Initialize development environment
make setup

# Start application (PostgreSQL + Go server)
make run

# Access at http://localhost:8080
```

### Production Deployment
```bash
# Setup production environment
make prod-setup

# Deploy with Docker Compose
./deployment.sh

# Check deployment status  
make prod-status
```

## Documentation

- **[API Reference](docs/API.md)** - Complete API documentation
- **[Deployment Guide](DEPLOYMENT.md)** - Production deployment instructions
- **[Architecture Overview](docs/architecture.md)** - System design and diagrams

## Configuration

## Configuration

### Core Configuration (`configs/config.toml`)

```toml
[server]
port = 8080

[database]
connection_string = "postgres://ryu:pass@localhost:5432/yoru_pastebin?sslmode=disable"

[paste]
id_length = 8
default_expiration_minutes = 1440  # 24 hours
max_content_size_bytes = 20971520  # 20MB

[security]
bcrypt_cost = 12
rate_limit_per_second = 2

[s3]
bucket = "your-s3-bucket-name"
region = "your-aws-region"
```

### Environment Variables (`.env`)

```bash
# Server Configuration
SERVER_PORT=8080
BASE_URL=https://paste.alokranjan.me
ENVIRONMENT=production

# Database
DATABASE_CONNECTION_STRING=postgres://yoru_user:${POSTGRES_PASSWORD}@db:5432/yoru_pastebin?sslmode=require
POSTGRES_USER=yoru_user
POSTGRES_PASSWORD=secure_password
POSTGRES_DB=yoru_pastebin

# AWS S3 Storage
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_S3_BUCKET=yoru-pastebin-files

# Security & Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=60

# TLS Configuration
TLS_CERT_EMAIL=your-email@domain.com
```

## Architecture
### Production Infrastructure
```mermaid
graph TB

    %% define graph nodes
    subgraph "Load Balancer & SSL"
        LB[Traefik v2.10<br/>SSL Termination<br/>Rate Limiting]
    end

    subgraph "Application Layer"
        LB --> APP1[Yoru 1<br/>Port 8080]
        LB --> APP2[Yoru 2<br/>Port 8080]
    end

    subgraph "Data Layer"
        APP1 --> DB[(PostgreSQL 16<br/>Persistent Storage)]
        APP2 --> DB
        APP1 --> FS[Hybrid File Storage<br/>Local or S3]
        APP2 --> FS
    end

    subgraph "Security"
        SEC1[CSP Headers]
        SEC2[Rate Limiting]
        SEC3[TLS 1.3]
    end

    %% assign classes
    classDef mocha fill:#1e1e2e,stroke:#313244,color:#cdd6f4;
    classDef blue fill:#1e1e2e,stroke:#89b4fa,color:#89b4fa;
    classDef pink fill:#1e1e2e,stroke:#f5c2e7,color:#f5c2e7;
    classDef green fill:#1e1e2e,stroke:#a6e3a1,color:#a6e3a1;
    classDef yellow fill:#1e1e2e,stroke:#f9e2af,color:#f9e2af;

    class LB,APP1,APP2 mocha;
    class DB yellow;
    class FS green;
    class SEC1,SEC2,SEC3 pink;
```

### Container Orchestration

**Development (`docker-compose.yml`):**
- Single replica for local development
- Direct database connection
- Volume mounting for live reload

**Production (`docker-compose.prod.yml`):**
- 2 replicas with load balancing
- External network configuration
- Health checks and restart policies
- Traefik integration with SSL

### Security Implementation

- **Traefik Middleware:** Security headers, rate limiting, SSL redirect
- **Database Security:** SSL connections, credential isolation
- **Container Security:** Non-root users, read-only filesystems
- **Network Security:** Internal Docker networks, isolated services

## Production Deployment

### Using Docker Compose

```bash
# 1. Clone and setup
git clone https://github.com/ryu-ryuk/yoru-pastebin.git
cd yoru-pastebin

# 2. Configure environment
cp .env.example .env
# Edit .env with your production values

# 3. Deploy with automated script
./deployment.sh

# 4. Monitor deployment
make prod-status
docker-compose -f docker-compose.prod.yml logs -f
```

### Manual Deployment Steps

```bash
# Build production image
docker build -t yoru-pastebin:latest .

# Create external network
docker network create yoru_web

# Deploy with production compose
docker-compose -f docker-compose.prod.yml up -d

# Check container health
docker-compose -f docker-compose.prod.yml ps
```

### Infrastructure Components

- **Application:** Go 1.24.4 multi-stage Docker build
- **Database:** PostgreSQL 16 Alpine with persistent volumes
- **Reverse Proxy:** Traefik v2.10 with Let's Encrypt integration
- **File Storage:** AWS S3 with configurable regions
- **Networking:** Docker overlay networks for service isolation

## API Usage

See [API Reference](docs/API.md) for complete documentation.

**Create a paste:**
```bash
curl -X POST https://paste.alokranjan.me/api/v1/pastes \
  -H "Content-Type: application/json" \
  -d '{
    "content": "package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello, Yoru!\")\n}",
    "language": "go",
    "expiration_minutes": 1440
  }'
```

**Retrieve a paste:**
```bash
curl https://paste.alokranjan.me/api/v1/pastes/aB3kX9mP
```

**Upload a file:**
```bash
curl -X POST https://paste.alokranjan.me/api/v1/pastes \
  -F "file=@example.go" \
  -F "language=go" \
  -F "expiration_minutes=60"
```

## Security Features

### Encryption Implementation
- **AES-256-GCM** for authenticated encryption
- **PBKDF2-SHA256** for password-based key derivation
- **Constant-time comparison** for password verification
- **Secure random ID generation** with sufficient entropy

## Deployment & Operations

### Health Monitoring
```bash
# Check application health
curl https://paste.alokranjan.me/health

# Monitor logs
docker-compose -f docker-compose.prod.yml logs -f yoru

# Check database status
docker-compose -f docker-compose.prod.yml exec db psql -U yoru_user -d yoru_pastebin -c "\l"
```

### Backup & Maintenance
```bash
# Manual database backup
docker-compose -f docker-compose.prod.yml exec db pg_dump -U yoru_user yoru_pastebin > backup.sql

# Update deployment
./deployment.sh

# Scale application instances
docker-compose -f docker-compose.prod.yml up -d --scale yoru=3
```

### Security Features

- **Traefik Security Headers:** Frame denial, XSS protection, HSTS
- **Rate Limiting:** 10 requests/minute average, 20 burst
- **Content Security Policy:** Strict policy for XSS prevention
- **TLS Configuration:** Modern TLS with automatic certificate renewal
- **Database Security:** SSL connections, isolated credentials

## Development

### Prerequisites
- Go 1.22+ (application uses Go 1.24.4 in Docker)
- Docker & Docker Compose
- PostgreSQL (for local development)
- Make (optional but recommended)

### Available Make Commands
```bash
make help           # Show all available commands
make setup          # Initialize development environment
make start_db       # Start PostgreSQL container
make run            # Start the application
make test           # Run test suite
make build          # Build production binary
make clean          # Clean up development environment
make prod-setup     # Setup production environment
make prod-deploy    # Deploy to production
make prod-status    # Check production deployment status
```

### Development Workflow
```bash
# First time setup
make setup          # Creates DB, runs migrations, builds app

# Daily development
make start_db       # Start database if not running
make run            # Start application with hot reload

# Testing and building
make test           # Run unit and integration tests
make build          # Create optimized production build
```

### Directory Structure
```
cmd/yoru/           # Application entry point
internal/           # Private application code
├── config/         # Configuration management
├── database/       # Database connection and utilities
├── paste/          # Core paste functionality
└── server/         # HTTP server and routing
pkg/                # Public packages
├── crypt/          # Encryption utilities
└── idgen/          # ID generation
web/                # Frontend assets
├── static/         # CSS, JS, images
└── templates/      # HTML templates
db/migrations/      # Database schema migrations
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run the test suite (`make test`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Submit a pull request

### Development Guidelines
- Follow Go best practices and `gofmt` formatting
- Include unit tests for new features
- Update documentation for API changes
- Test with both development and production Docker configurations

### Testing [TODO]
```bash
# Run all tests
make test

# Run specific test packages
go test ./internal/paste/...
go test ./pkg/crypt/...
```

## License

Licensed under the GNU General Public License v3.0. See [LICENSE](LICENSE) for details.

---

**Built for developers who prioritize security, performance and a whole lot of catppuccin ;) **
