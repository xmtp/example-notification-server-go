# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

XMTP push notification server in Go. Designed to be forked and customized per application. Receives messages from the XMTP network and delivers push notifications via APNS, FCM, or HTTP webhooks.

## Common Commands

### Build & Run
```bash
./dev/up              # Start Docker services (PostgreSQL, XMTP node) — run once
./dev/build           # Build binary to ./dist/
./dev/start           # Run server with both API and listener enabled
./dev/run --api       # Run API server only
./dev/run --help      # Show all CLI options
```

### Testing
```bash
go test -p 1 ./...    # Unit tests (must run serially — shared database)
./dev/integration     # Integration tests (Docker-based, TypeScript/Bun)
```

Tests require `./dev/up` to be running. The `-p 1` flag is required because tests share a database instance that gets wiped between tests.

### Linting
```bash
golangci-lint run --config dev/.golangci.yaml   # Go linter
./dev/lint-shellcheck                            # Shell script linter
```

### Proto Generation
```bash
./dev/gen-proto       # Regenerate all proto code (requires buf CLI)
buf generate          # Regenerate notification protos only
```

## Architecture

### Two-Service Design

The server runs two optional components via flags:
- **API server** (`--api`): Connect RPC over HTTP on port 8080. Handles device registration and topic subscriptions. Stateless, can scale horizontally.
- **XMTP listener** (`--xmtp-listener`): Persistent gRPC stream to XMTP node. Receives messages, looks up subscriptions, delivers notifications. Should be a single instance.

Both are started as goroutines from `cmd/server/main.go` and shut down gracefully on SIGINT/SIGTERM.

### Package Structure

- `cmd/server/main.go` — Entry point, wires up dependencies, starts services
- `pkg/api/` — Connect RPC API server (registration, subscribe/unsubscribe endpoints)
- `pkg/xmtp/` — XMTP network listener with worker pool (default 50 workers, channel capacity 100)
- `pkg/delivery/` — Push notification delivery: APNS (`apns.go`), FCM (`fcm.go`), HTTP (`http.go`)
- `pkg/installations/` — Installation CRUD (device registration)
- `pkg/subscriptions/` — Subscription management (topic subscriptions with optional HMAC keys)
- `pkg/interfaces/` — Core interfaces (`Installations`, `Subscriptions`, `Delivery`) and domain types
- `pkg/db/` — Bun ORM models and migrations (PostgreSQL)
- `pkg/topics/` — Topic parsing and message type detection (V3 MLS topics only)
- `pkg/options/` — CLI flag/env var configuration structs
- `pkg/proto/` — Generated protobuf/Connect code (~28K LOC, do not edit)

### Key Patterns

- **Dependency injection**: Services receive interfaces via constructors; `main.go` wires the object graph
- **Interface-driven**: Core abstractions in `pkg/interfaces/`; mocks generated via `mockery` into `mocks/`
- **Strategy pattern for delivery**: Multiple `Delivery` implementations; listener checks `CanDeliver()` then delegates
- **Soft deletes**: Installations use `deleted_at` field
- **HMAC sender filtering**: Prevents self-notifications using 30-day rolling keys

### Configuration

Options are parsed from CLI flags and environment variables (via `go-flags` struct tags). Local dev uses `.env.local`, Docker uses `.env.docker`.

### Database

PostgreSQL via Bun ORM. Migrations in `pkg/db/migrations/`. Test DSN: `postgres://postgres:xmtp@localhost:25432/postgres?sslmode=disable`.

### Testing Approach

- Unit tests use mockery-generated mocks for interface boundaries
- API tests start a real HTTP server with mocked services
- `test/helpers.go` provides `CreateTestDb()` with automatic table truncation
- Integration tests run full stack in Docker with HTTP delivery for verification
