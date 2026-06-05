# WarSpark

[简体中文](README.zh-CN.md)

Reusable Go REST API backend template built with Gin, Cobra, Viper, MySQL, Redis, JWT, Snowflake IDs, slog, Swagger, and golang-migrate.

## What Is Included

- HTTP server with graceful shutdown/restart support
- Cobra CLI with `server` and `migrate` commands
- YAML, environment variable, and CLI override configuration
- MySQL and Redis infrastructure initialization
- JWT manager and auth middleware
- Request ID, recovery, access logging, and CORS middleware
- Unified JSON responses, error mapping, pagination, and request validation helpers
- Health and readiness endpoints
- Swagger generation and `/swagger/*any`
- Database migrations under `migrations/`
- Makefile and GitHub Actions checks
- `docs/AI_DEVELOPMENT.md` for AI coding tools

## Initialize A New Project

macOS/Linux:

```bash
bash scripts/init-template.sh \
  --module-path github.com/acme/demo-api \
  --slug demo-api \
  --display-name "Demo API"
```

Windows PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/init-template.ps1 `
  -ModulePath github.com/acme/demo-api `
  -Slug demo-api `
  -DisplayName "Demo API"
```

Then refresh generated files:

```bash
go mod tidy
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/demo-api/main.go -o docs
go test -buildvcs=false ./...
```

## Run

```bash
go run ./cmd/warspark server --config=configs/config.dev.yaml
```

Default endpoints:

- `GET /health`
- `GET /api/v1/health`
- `GET /api/v1/ready`
- `GET /api/v1`
- `GET /swagger/index.html`

## Configuration

Configuration precedence is:

```text
defaults -> YAML file -> WARSPARK_* environment variables -> CLI flags
```

Examples:

- `WARSPARK_JWT_SECRET`
- `WARSPARK_MYSQL_HOST`
- `WARSPARK_SERVER_PORT`

Production secrets should come from environment variables or a secret manager, not committed YAML.
Database names and other deployment-specific values are configured through YAML or environment variables; the init scripts only update sample defaults for convenience.

## Common Commands

```bash
make run
make test
make build
make check
make docs
make migrate-create MIGRATION_NAME=create_users
make migrate-up
make migrate-down STEPS=1
make migrate-version
```

## Structure

```text
cmd -> app -> api -> controller -> service -> repository -> model
```

Keep command parsing in `cmd`, startup and orchestration in `internal/app`, HTTP binding in `internal/controller`, business logic in `internal/service`, persistence in `internal/repository`, and table structs in `internal/model`.

See [docs/design.md](docs/design.md), [docs/template-spec.md](docs/template-spec.md), and [docs/AI_DEVELOPMENT.md](docs/AI_DEVELOPMENT.md).
