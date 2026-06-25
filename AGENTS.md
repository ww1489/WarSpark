# WarSpark — AGENTS.md

## Global Rules

- 使用中文回复。所有交互、代码注释、文档说明等均使用简体中文。

Go REST API backend template. Module: `github.com/ww1489/WarSpark`, Go 1.26.3.

## Commands

| `make run` | `go run ./cmd/warspark server --config=configs/config.dev.yaml` |
| `make test` | `go test -buildvcs=false ./...` |
| `make check` | fmt → vet → test → build (standard quality gate) |
| `make fmt` | `gofmt -w .` |
| `make docs` | Regenerate Swagger from `cmd/warspark/main.go` annotations |
| `make migrate-create MIGRATION_NAME=x` | Create sequential `.up.sql`/`.down.sql` pair |
| `make migrate-up` / `make migrate-down STEPS=1` | Apply/revert migrations via Cobra |
| `go run ./cmd/warspark cocapi syncapi --token=xxx` | Fetch latest CoC swagger + regenerate client code (recommended) |
| `go run ./cmd/warspark cocapi fetch --token=xxx` | Fetch latest CoC swagger only |
| `go run ./cmd/warspark cocapi generate` | Regenerate CoC client code from existing swagger |

CI runs: formatting check (`gofmt -l .` → no output), `go vet ./...`, `go test -buildvcs=false ./...`, `go build -buildvcs=false ./cmd/warspark`.

## Architecture

Control flow: `cmd` (Cobra entry) → `internal/app` (orchestration) → `internal/api/v1` (route setup) → `internal/controller` (input binding) → `internal/service` (business logic) → `internal/repository` (SQL/Redis).

- **No DI container.** Explicit constructors, explicit dependency passing (see `internal/app/bootstrap.go:87-105` for the wiring pattern).
- **No `internal/model` yet** — domain structs live in `internal/domain/` (war, imagesearch, layout) alongside app-level DTOs.
- New resource order: model → migration → repository → service → controller → routes (registered in `internal/api/v1/routes.go`) → tests → Swagger annotations.

## Config

Viper with precedence: defaults → YAML → `WARSPARK_*` env vars → CLI flags. Config struct at `internal/config/config.go`. Add new fields to `Config`, `Defaults()`, `setDefaults()`, `Validate()`, and update `configs/config.*.yaml`.

## Key Conventions

- **Response shape**: `{"code": 0, "message": "ok", "data": {}}`. Use `utils.OK()` / `utils.Fail()`.
- **Errors**: Domain errors use `utils.AppError` with stable codes (`ErrNotFound`, `ErrUnauthorized`, etc.) and HTTP mapping via `StatusForCode()`.
- **Transactions**: Use `inframysql.WithinTx(ctx, db, func(tx *sqlx.Tx) error)` in services, not controllers.
- **IDs**: Snowflake (from `pkg/snowflake`) or UUID (see migration PKs — both patterns exist).
- **Auth**: JWT via `Authorization: Bearer <token>` header. Use `auth.Required(mgr)` middleware for admin routes. Read claims via `auth.UserID(c)`, `auth.UserRole(c)`, `auth.Claims(c)`.
- **Validation**: Gin binding tags + `utils.BindJSON`/`utils.BindQuery` in controllers.
- **Logging**: `slog.Logger` passed through `RuntimeConfig`. File rotation via lumberjack.
- **Swagger**: Enabled/disabled via `docs.swagger_enabled` config. Generated from Gin handler annotations.
- **Workers**: Background workers start in `internal/app/bootstrap.go` (see `imageSearchWorker.Start(appCtx)` pattern). Accept `context.Context` for cancellation.

## Template Info

This is a reusable template. Init scripts: `scripts/init-template.sh` (macOS/Linux) and `scripts/init-template.ps1` (Windows). Updating template metadata (app name, module path, display name) must keep both scripts behaviorally equivalent.

## Verified Docs Sources

For AI tool guidance, `docs/AI_DEVELOPMENT.md` is the authoritative local reference — it covers adding resources, error handling, validation, transactions, auth context, config changes, migrations, and verification. This file (AGENTS.md) distills only what `docs/AI_DEVELOPMENT.md` leaves implicit or that differs from Go conventions.
