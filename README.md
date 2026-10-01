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
Client / API Consumer
        |
        v
  Fiber HTTP API
        |
        v
  Service Layer
        |
        +----------------+
        |                |
        v                v
 Repository Layer   gRPC Clients
        |                |
        v                v
    Database       Internal Services
                         |
                         v
                    gRPC Servers
```

## Current Services

- **user-service**: Handles user-related operations and exposes gRPC handlers backed by service and repository layers.

## Tech Stack

- **Go** for backend development
- **Fiber** for HTTP endpoints
- **gRPC** for internal communication
- **Protocol Buffers** for API contracts
- **MongoDB** for persistence
- **Docker** for local development support

## Goal

The goal of this project is to demonstrate how task management functionality can be built with a clean Go backend architecture where external requests come through Fiber and internal backend communication is handled through gRPC.
