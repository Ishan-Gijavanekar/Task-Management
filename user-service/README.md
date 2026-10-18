# Go gRPC User Service

A production-oriented backend service built with **Go**, **gRPC**, **Fiber**, and **MongoDB**.

The project demonstrates how to build a scalable backend using clean separation between transport, business logic, and persistence layers.

## Tech Stack

- **Go** — Backend language
- **gRPC** — Primary service communication
- **Protocol Buffers** — API contracts and serialization
- **Fiber** — HTTP/REST API layer
- **MongoDB** — Database
- **Docker** — Local development and deployment

## Architecture

```text
          Client
             │
      ┌──────┴──────┐
      │             │
     REST          gRPC
      │             │
      ▼             ▼
   Fiber         gRPC Server
      │             │
      └──────┬──────┘
             ▼
       Service Layer
             │
             ▼
      Repository Layer
             │
             ▼
          MongoDB
```

## Project Structure

```text
.
├── api/                # Protocol Buffer definitions
├── cmd/
│   └── server/         # Application entry point
├── gen/                # Generated protobuf/gRPC code
├── internal/
│   ├── config/         # Application configuration
│   ├── database/       # Database initialization
│   ├── domain/         # Domain models
│   ├── repository/     # Data access layer
│   ├── service/        # Business logic
│   └── transport/      # gRPC and HTTP handlers
├── pkg/                # Shared packages
├── tests/              # Integration and E2E tests
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── go.mod
```

## Getting Started

### Prerequisites

Make sure you have installed:

- Go
- Docker
- Docker Compose

### Setup

Clone the repository and install dependencies:

```bash
go mod download
```

Create the environment configuration:

```bash
cp .env.example .env
```

Start MongoDB:

```bash
docker compose up -d
```

Run the application:

```bash
go run ./cmd/server
```

## Default Ports

| Service | Port |
|---|---:|
| HTTP / Fiber | `8080` |
| gRPC | `50051` |
| MongoDB | `27017` |

## Development Status

🚧 **Under active development**

Planned functionality includes:

- User CRUD operations
- gRPC API
- REST API with Fiber
- MongoDB persistence
- Authentication and authorization
- gRPC interceptors
- Request validation
- Pagination and filtering
- Streaming RPCs
- Structured logging
- Health checks
- Unit and integration tests
- Dockerized deployment

## License

This project is intended for learning and demonstration purposes.

## Testing and quality

Run the standard checks from this directory:

```sh
make quality
```

Run race detection and produce a coverage report with:

```sh
make test-race
make test-cover
```

The MongoDB repository test is opt-in so the normal unit suite remains self-contained. Start MongoDB, then run:

```sh
MONGO_INTEGRATION_URI='mongodb://root:rootpassword@localhost:27017/?authSource=admin' make test-integration
```

## Authentication lifecycle

The REST API supports short-lived access tokens and rotating refresh sessions:

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/reset-password`
- `POST /api/v1/auth/verify-email`
- `POST /api/v1/auth/resend-verification`

Refresh tokens are stored as SHA-256 hashes and sent to browsers using an HttpOnly cookie. Reusing a rotated refresh token revokes its token family. Set `AUTH_REQUIRE_EMAIL_VERIFICATION=true` to require verification before login. Development verification and reset links are written to structured logs; non-development environments should configure the SMTP variables from `.env.example`.

`GET /health/live` reports process liveness. `GET /health/ready` checks MongoDB with a bounded timeout. Keep the compatibility endpoint at `GET /health` for existing probes.
