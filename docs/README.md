# API Documentation

## Overview

The ME CLI Sunset API provides a RESTful interface for managing XL Axiata mobile accounts, data packages, family plans, and circle groups.

## OpenAPI Specification

The full API specification is available in [OpenAPI 3.0 format](../api/openapi.yaml).

You can view it using any OpenAPI-compatible tool:

- [Swagger Editor](https://editor.swagger.io/) — Paste the YAML content
- [Swagger UI](https://swagger.io/tools/swagger-ui/) — Load the specification file
- [Redoc](https://redocly.com/) — Generate beautiful documentation

## Quick Start

### 1. Start the Server

```bash
cp .env.template .env
# Edit .env with your credentials
go build -o me-cli-sunset ./cmd/server
./me-cli-sunset
```

### 2. Check Health

```bash
curl http://localhost:8080/api/v1/health
```

### 3. Authenticate

```bash
# Request OTP
curl -X POST http://localhost:8080/api/v1/auth/otp/request \
  -H "Content-Type: application/json" \
  -d '{"contact": "6281234567890"}'

# Submit OTP
curl -X POST http://localhost:8080/api/v1/auth/otp/submit \
  -H "Content-Type: application/json" \
  -d '{"contact": "6281234567890", "code": "123456"}'
```

### 4. Browse Packages

```bash
# Get balance
curl http://localhost:8080/api/v1/packages/balance

# Get package families
curl "http://localhost:8080/api/v1/packages/families?category_code=DATA"

# Get package detail
curl -X POST http://localhost:8080/api/v1/packages/detail \
  -H "Content-Type: application/json" \
  -d '{"package_option_code": "PKG001"}'
```

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Client    │────▶│  API Server  │────▶│  Upstream   │
│  (HTTP)     │◀────│  (Go)        │◀────│  API (XL)   │
└─────────────┘     └──────┬───────┘     └─────────────┘
                           │
                    ┌──────▼───────┐
                    │   SQLite DB  │
                    │  (accounts,  │
                    │  bookmarks,  │
                    │  fingerprints)│
                    └──────────────┘
```

## Repository

- **Source**: [github.com/kertasbaru/me-cli-sunset](https://github.com/kertasbaru/me-cli-sunset)
