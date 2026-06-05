# Template Spec

## Default Identity

| Item | Default |
| --- | --- |
| Module path | `github.com/ww1489/WarSpark` |
| Command | `warspark` |
| Display name | `WarSpark` |
| Environment prefix | `WARSPARK` |

Use `scripts/init-template.sh` on macOS/Linux or `scripts/init-template.ps1` on Windows PowerShell to initialize these values for a real project.

## Required Template Guarantees

- The repository must compile before initialization.
- The repository must compile after initialization with a valid module path and slug.
- No domain-specific business model should exist in the default template.
- Common infrastructure must be present but not force a heavy framework shape.
- The existing directory structure should stay stable unless repeated real code justifies a split.

## Initialization Contract

Required parameters:

- `ModulePath`: Go module path, for example `github.com/acme/demo-api`.
- `Slug`: command and service slug, for example `demo-api`.
- `DisplayName`: human-facing service name, for example `Demo API`.

Optional parameters:

- `EnvPrefix`: defaults to uppercase slug with hyphens converted to underscores.

## Runtime Config Boundary

Initialization parameters are template identity only. Do not add init flags for database names, passwords, hosts, Redis settings, JWT secrets, log paths, or Swagger exposure.

The init scripts may update derived sample defaults such as `mysql.user`, `mysql.db`, `jwt.issuer`, and `log.file.path` so a generated project is coherent out of the box. Actual runtime values remain controlled by YAML files, environment variables, CLI overrides, or deployment secrets.

After initialization, run:

```bash
go mod tidy
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/<slug>/main.go -o docs
go test -buildvcs=false ./...
```

Initialization examples:

```bash
bash scripts/init-template.sh \
  --module-path github.com/acme/demo-api \
  --slug demo-api \
  --display-name "Demo API"
```

```powershell
powershell -ExecutionPolicy Bypass -File scripts/init-template.ps1 `
  -ModulePath github.com/acme/demo-api `
  -Slug demo-api `
  -DisplayName "Demo API"
```

## Common Extension Points

- Add HTTP resources in `internal/controller`, `internal/service`, `internal/repository`, and `internal/model`.
- Add new route groups in `internal/api/v1/routes.go`.
- Add new config fields in `internal/config/config.go`, then update config YAML and tests.
- Add migrations under `migrations/`.
- Add reusable business-neutral utilities in `internal/utils`.
- Add infrastructure adapters under existing `internal/infra/<adapter>` packages.

## Acceptance Checks

Run these before publishing a template update:

```bash
gofmt -w .
go vet ./...
go test -buildvcs=false ./...
go build -buildvcs=false ./cmd/warspark
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/warspark/main.go -o docs
```

Search checks:

```bash
rg "ClanChan|clanchan|CLANCHAN|贴吧|吧主|帖子|回复" -g "!docs/template-spec.md"
```

The search should return no project code or documentation references to old domain content.
