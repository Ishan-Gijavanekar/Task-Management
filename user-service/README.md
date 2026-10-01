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