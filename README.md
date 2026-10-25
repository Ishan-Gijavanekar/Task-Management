# Task Management gRPC

Task Management gRPC is a Go backend project for handling task-related workflows with a service-oriented design. The project uses **Fiber** for external HTTP endpoints and **gRPC** for internal service-to-service communication, keeping public API access separate from fast internal calls.

## Overview

The application is intended to manage task handling through clear layers:

- **Fiber HTTP layer** exposes REST-style endpoints for clients.
- **gRPC layer** handles internal calls between backend services.
- **Service layer** contains business rules for task and related domain operations.
- **Repository layer** manages persistence and database access.
- **Protocol Buffers** define typed contracts for gRPC communication.

This structure makes it easier to grow the project into multiple services while keeping internal communication strongly typed and efficient.

## Architecture

```text
Browser / API Consumer
        |
        v
 API Gateway :8088
        |
        +-------------------+
        |                   |
        v                   v
 User Service         Task Service
        |                   |
        +---------+---------+
                  |
                  v
               MongoDB
```

## Current Services

- **user-service**: Handles user-related operations and exposes gRPC handlers backed by service and repository layers.
- **task-service**: Handles workspace-scoped task lifecycle, filtering, ordering, idempotency, activity history, REST, and gRPC APIs.
- **gateway-service**: Exposes the single browser API, validates authentication, normalizes errors, enforces HTTP limits, and publishes the OpenAPI contract and generated frontend types.

Run the complete local stack with `docker compose up --build`. Use `http://localhost:8088/api/v1` as the public API origin and `http://localhost:8088/openapi.json` for its contract. The user and task HTTP ports remain internal to the Compose network.

## Tech Stack

- **Go** for backend development
- **Fiber** for HTTP endpoints
- **gRPC** for internal communication
- **Protocol Buffers** for API contracts
- **MongoDB** for persistence
- **Docker** for local development support

## Goal

The goal of this project is to demonstrate how task management functionality can be built with a clean Go backend architecture where external requests come through Fiber and internal backend communication is handled through gRPC.
