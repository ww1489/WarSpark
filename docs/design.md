# WarSpark Design

> Version: v1
> Updated: 2026-06-03
> Module: `github.com/ww1489/WarSpark`

## Purpose

This repository is a reusable Go REST API backend template. It provides the infrastructure and conventions that most service projects need before domain code is added.

The template is intentionally a modular monolith. It should stay simple until real business complexity requires a split.

## Architecture

Dependency direction:

```text
cmd -> app -> api -> controller -> service -> repository -> model
              -> middleware, utils, pkg
```

Responsibilities:

| Layer | Responsibility |
| --- | --- |
| `cmd/<app>` | Cobra commands, flags, help text |
| `internal/app` | Startup, shutdown, infrastructure wiring, migration execution |
| `internal/api` | Route registration and API version grouping |
| `internal/controller` | HTTP binding, validation, service calls, response conversion |
| `internal/service` | Business rules, transactions, authorization checks |
| `internal/repository` | SQL and Redis access, no business policy |
| `internal/model` | Storage-facing structs |
| `internal/middleware` | Request ID, recovery, logging, CORS, auth |
| `internal/infra` | MySQL, Redis, logger adapters |
| `internal/utils` | Response, errors, pagination, validation helpers |
| `pkg` | Reusable technical packages such as JWT and Snowflake |

## Runtime Flow

1. `cmd/<app>` parses CLI flags.
2. `internal/config.Load` resolves defaults, YAML, environment variables, then CLI overrides.
3. `internal/app.Start` initializes logger, MySQL, Redis, JWT manager, Gin router, and middleware.
4. `internal/api/v1.SetupRoutes` registers public routes.
5. The platform starts the HTTP server and handles shutdown.

Configuration precedence:

```text
defaults -> YAML -> WARSPARK_* environment variables -> CLI flags
```

## Included Capabilities

- Gin HTTP API
- Cobra CLI
- Viper configuration
- MySQL via sqlx
- Redis via go-redis
- JWT access and refresh token utilities
- Snowflake ID generator
- slog console and rotating file logs
- Request ID, recovery, access logging, CORS, auth middleware
- Health and readiness checks
- Unified response and error helpers
- Request validation helpers
- Pagination helpers
- Database migration command using golang-migrate
- Swagger generation and route

## API Conventions

Base path:

```text
/api/v1
```

Response shape:

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

Error code ranges:

| Range | Meaning |
| --- | --- |
| `1xxxx` | Generic framework and platform errors |
| `2xxxx+` | Reserved for application domains |

Default generic codes live in `internal/utils/errors.go`.

## Configuration

Default config files:

- `configs/config.example.yaml`
- `configs/config.dev.yaml`
- `configs/config.prod.yaml`

Sensitive values such as `jwt.secret`, database passwords, and Redis passwords must be injected by environment variables or a secret manager in production.
Deployment-specific values such as database names, hosts, Redis settings, JWT secrets, log paths, and Swagger exposure stay in runtime config rather than template initialization flags.

## Database Migrations

Migration files live in `migrations/` and use sequential naming:

```text
000001_create_users.up.sql
000001_create_users.down.sql
```

Commands:

```bash
go run ./cmd/warspark migrate up --config=configs/config.dev.yaml
go run ./cmd/warspark migrate down --steps=1 --config=configs/config.dev.yaml
go run ./cmd/warspark migrate version --config=configs/config.dev.yaml
```

## Development Rules

- Keep business logic out of `cmd`, route setup, middleware, and controllers.
- Prefer explicit constructors and direct dependency passing.
- Do not add a new top-level or `internal/*` directory unless a repeated pattern justifies it.
- Put errors, response helpers, validation, and pagination in `internal/utils`.
- Put MySQL transaction helpers in `internal/infra/mysql`.
- Put migration execution in `internal/app`.
- Keep docs synchronized when structure, commands, config, or public API changes.
