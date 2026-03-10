# ME CLI Sunset API

> REST API service for managing XL Axiata mobile accounts, data packages, family plans, and circle groups — rewritten in Go with SQLite credential storage.

**Repository**: [github.com/kertasbaru/me-cli-sunset](https://github.com/kertasbaru/me-cli-sunset)

## Overview

This project is a professional API server that replaces the original Python CLI application. It provides a RESTful interface for all XL ME (MYnyak Engsel) operations with:

- **Go (Golang)** — type-safe, compiled, high-performance backend
- **SQLite** — local database for credential and session storage (replaces JSON files)
- **OpenAPI 3.0** — professional API documentation
- **Standard project layout** — following Go community conventions

## Project Structure

```
├── cmd/server/             # Application entry point
│   └── main.go
├── internal/
│   ├── config/             # Environment configuration
│   ├── database/           # SQLite database layer
│   ├── model/              # Data models and request/response types
│   ├── crypto/             # AES encryption, HMAC signatures, fingerprinting
│   ├── client/             # Upstream API client (engsel, ciam, circle, famplan)
│   ├── handler/            # HTTP request handlers
│   ├── middleware/          # HTTP middleware (logging, CORS, recovery)
│   └── service/            # Business logic services (auth, tokens)
├── api/
│   └── openapi.yaml        # OpenAPI 3.0 specification
├── docs/                   # Additional documentation
├── go.mod                  # Go module definition
└── .env.template           # Environment variable template
```

## Getting Started

### Prerequisites

- Go 1.21 or later
- Environment variables (see `.env.template`)

### Installation

```bash
# Clone the repository
git clone https://github.com/kertasbaru/me-cli-sunset.git
cd me-cli-sunset

# Copy environment template and fill in values
cp .env.template .env

# Install dependencies
go mod tidy

# Build the server
go build -o me-cli-sunset ./cmd/server

# Run the server
./me-cli-sunset
```

### Configuration

Copy `.env.template` to `.env` and set the required values:

| Variable | Description | Required |
|---|---|---|
| `SERVER_PORT` | HTTP server port (default: 8080) | No |
| `DATABASE_PATH` | SQLite database path (default: me_cli.db) | No |
| `BASE_API_URL` | Base API URL | Yes |
| `BASE_CIAM_URL` | Base CIAM authentication URL | Yes |
| `BASIC_AUTH` | Basic auth credentials | Yes |
| `AX_FP_KEY` | Device fingerprint key | Yes |
| `UA` | User-Agent string | Yes |
| `API_KEY` | API key | Yes |
| `ENCRYPTED_FIELD_KEY` | Encrypted field key | Yes |
| `XDATA_KEY` | XData encryption key | Yes |
| `AX_API_SIG_KEY` | AX API signature key | Yes |
| `X_API_BASE_SECRET` | X API base secret | Yes |
| `CIRCLE_MSISDN_KEY` | Circle MSISDN encryption key | No |

## API Documentation

Full API documentation is available in OpenAPI 3.0 format at [`api/openapi.yaml`](api/openapi.yaml).

### Endpoints Overview

#### Health
| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/health` | Server health check |

#### Authentication
| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/auth/otp/request` | Request OTP via SMS |
| `POST` | `/api/v1/auth/otp/submit` | Submit OTP and authenticate |
| `GET` | `/api/v1/auth/accounts` | List all stored accounts |
| `POST` | `/api/v1/auth/accounts/switch` | Switch active account |
| `DELETE` | `/api/v1/auth/accounts/{number}` | Delete a stored account |
| `GET` | `/api/v1/auth/active` | Get active user info |
| `POST` | `/api/v1/auth/renew` | Force token renewal |

#### Packages
| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/packages/balance` | Get balance and credit |
| `POST` | `/api/v1/packages/family` | Get package family options |
| `GET` | `/api/v1/packages/families` | List package families |
| `POST` | `/api/v1/packages/detail` | Get package detail |
| `GET` | `/api/v1/packages/addons` | Get package addons |
| `GET` | `/api/v1/packages/tiering` | Get loyalty tiering info |
| `GET` | `/api/v1/packages/quotas` | Get quota details |
| `POST` | `/api/v1/packages/payment-methods` | Get payment methods |
| `POST` | `/api/v1/packages/unsubscribe` | Unsubscribe from package |

#### Circle (Family Hub)
| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/circle/groups` | Get circle groups |
| `POST` | `/api/v1/circle/groups` | Create new circle |
| `GET` | `/api/v1/circle/groups/{group_id}/members` | Get group members |
| `POST` | `/api/v1/circle/members/validate` | Validate member MSISDN |
| `POST` | `/api/v1/circle/members/invite` | Invite member |
| `POST` | `/api/v1/circle/members/remove` | Remove member |
| `POST` | `/api/v1/circle/invitations/accept` | Accept invitation |
| `POST` | `/api/v1/circle/spending-tracker` | Get spending tracker |
| `POST` | `/api/v1/circle/bonus` | Get bonus data |

#### Family Plan
| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/family/plan` | Get family plan data |
| `POST` | `/api/v1/family/validate` | Validate MSISDN |
| `POST` | `/api/v1/family/members/change` | Change member |
| `POST` | `/api/v1/family/members/remove` | Remove member |
| `POST` | `/api/v1/family/quota` | Set quota limit |

#### Notifications & Transactions
| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/notifications` | Get notifications |
| `GET` | `/api/v1/notifications/{notification_id}` | Get notification detail |
| `GET` | `/api/v1/transactions` | Get transaction history |
| `GET` | `/api/v1/dashboard` | Get dashboard data |

#### Bookmarks
| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/bookmarks` | List bookmarks |
| `POST` | `/api/v1/bookmarks` | Add bookmark |
| `DELETE` | `/api/v1/bookmarks/{id}` | Delete bookmark |

## Database Schema

The SQLite database (`me_cli.db`) contains the following tables:

### `accounts`
Stores user account credentials (replaces `refresh-tokens.json` and `active.number`).

| Column | Type | Description |
|---|---|---|
| `id` | INTEGER | Primary key |
| `number` | INTEGER | Phone number (unique) |
| `subscriber_id` | TEXT | Subscriber identifier |
| `subscription_type` | TEXT | Subscription type (PREPAID, PRIORITAS, etc.) |
| `refresh_token` | TEXT | OAuth refresh token |
| `is_active` | INTEGER | Whether this is the active account |

### `bookmarks`
Stores saved package bookmarks (replaces `bookmark.json`).

| Column | Type | Description |
|---|---|---|
| `id` | INTEGER | Primary key |
| `account_id` | INTEGER | Foreign key to accounts |
| `family_code` | TEXT | Package family code |
| `family_name` | TEXT | Package family name |
| `variant_name` | TEXT | Package variant name |
| `option_name` | TEXT | Package option name |

### `device_fingerprints`
Stores generated device fingerprints (replaces `ax.fp`).

| Column | Type | Description |
|---|---|---|
| `id` | INTEGER | Primary key |
| `fingerprint` | TEXT | AES-encrypted fingerprint |
| `device_id` | TEXT | MD5-derived device ID |

### `sentry_logs`
Stores quota monitoring log entries (replaces `sentry/*.jsonl`).

| Column | Type | Description |
|---|---|---|
| `id` | INTEGER | Primary key |
| `account_id` | INTEGER | Foreign key to accounts |
| `quotas` | TEXT | JSON quota data |
| `logged_at` | DATETIME | Timestamp |

## Migration from Python CLI

This Go API replaces the following Python components:

| Python Component | Go Replacement |
|---|---|
| `main.py` (CLI menus) | `cmd/server/main.go` (HTTP server) |
| `app/client/*.py` | `internal/client/` |
| `app/service/auth.py` + JSON files | `internal/service/auth.go` + SQLite |
| `app/service/crypto_helper.py` | `internal/crypto/crypto.go` |
| `app/client/encrypt.py` | `internal/crypto/crypto.go` |
| `app/service/bookmark.py` + JSON file | `internal/handler/bookmark.go` + SQLite |
| `app/menus/*.py` | `internal/handler/` (REST endpoints) |

## Terms of Service

By using this tool, the user agrees to comply with all applicable laws and regulations and to release the developer from any and all claims arising from its use.

## Credits

Originally created by [Purple Mashu](https://github.com/purplemashu) — [contact@mashu.lol](mailto:contact@mashu.lol)

Restructured as a professional Go API with SQLite storage.
