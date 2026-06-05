# AI Development Guide

This file is for AI coding tools working in this repository. Follow it before making code changes.

## Core Rules

- Keep the current directory structure. Do not add `internal/project`, `internal/apperror`, `internal/testutil`, or similar concept-only directories.
- For project initialization, use `scripts/init-template.sh` on macOS/Linux and `scripts/init-template.ps1` on Windows PowerShell. Keep both scripts behaviorally equivalent when changing template metadata.
- Use the existing dependency direction:

```text
cmd -> app -> api -> controller -> service -> repository -> model
```

- Do not put business logic in `cmd`, route registration, middleware, or controllers.
- Keep command handlers thin. They parse flags and call `internal/app`.
- Keep controllers thin. They bind input, validate input, call services, and convert responses.
- Keep repositories policy-free. They only access SQL or Redis.
- Prefer explicit constructors and explicit dependency passing. Do not introduce Wire, Fx, or a service container unless the repository has repeated wiring complexity.

## Adding A New Resource

For a resource such as `user`, `order`, or `project`, add code in this order:

1. Add storage structs in `internal/model` if tables are involved.
2. Add SQL migration files in `migrations/`.
3. Add repository methods in `internal/repository`.
4. Add business methods in `internal/service`.
5. Add request and response DTOs near the controller, or in `internal/controller/dto` if the package already exists.
6. Add controller methods in `internal/controller`.
7. Register routes in `internal/api/v1/routes.go`.
8. Add tests close to the changed package.
9. Add Swagger annotations for public endpoints.
10. Update README or docs when commands, config, routes, or structure change.

## Errors And Responses

- Use `internal/utils/errors.go` for shared error codes and mappings.
- Return `utils.AppError` or wrapped errors from services when the caller needs a stable API error.
- Use `utils.OK`, `utils.JSON`, or `utils.Fail` in controllers.
- Keep response shape stable:

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

## Validation

- Use Gin binding tags on request DTOs.
- Use `utils.BindJSON`, `utils.BindQuery`, or `utils.Validate` in controllers.
- Return field-level validation errors through `utils.ValidationFailed`.
- Add custom validators through Gin's validator engine from `internal/utils` or the package that owns the DTO.

## Transactions

- Use `internal/infra/mysql.WithinTx` for MySQL transactions.
- Keep transaction boundaries in services.
- Do not open transactions in controllers.

## Authentication Context

- Use helpers in `internal/middleware/auth` to read authenticated user data from Gin context.
- Do not read raw context keys outside the auth package unless a new helper is missing.

## Configuration

- Add config fields to `internal/config.Config`.
- Add defaults in `Defaults`.
- Add Viper defaults in `setDefaults`.
- Add validation in `Validate` when a value is required.
- Update `configs/config.example.yaml`, `configs/config.dev.yaml`, and `configs/config.prod.yaml`.
- Add or update config tests.
- Treat database names as runtime config (`mysql.db`), not as template initialization parameters.
- Do not add init flags for passwords, hosts, Redis settings, JWT secrets, log paths, or Swagger exposure; those belong in config.

## Migrations

- Use sequential files:

```text
000001_create_users.up.sql
000001_create_users.down.sql
```

- Use `make migrate-create MIGRATION_NAME=create_users`.
- Prefer reversible migrations.
- Do not edit migrations that may already have been applied in shared environments; add a new migration instead.

## Tests And Verification

Run before finishing:

```bash
gofmt -w .
go vet ./...
go test -buildvcs=false ./...
go build -buildvcs=false ./cmd/warspark
```

If Swagger annotations changed, also run:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/warspark/main.go -o docs
```
