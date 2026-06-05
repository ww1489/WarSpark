# WarSpark

[English](README.md)

基于 Gin、Cobra、Viper、MySQL、Redis、JWT、Snowflake ID、slog、Swagger 和 golang-migrate 的可复用 Go REST API 后端模板。

## 包含内容

- 支持优雅关闭/重启的 HTTP 服务
- 带 `server` 和 `migrate` 子命令的 Cobra CLI
- YAML、环境变量和 CLI 覆盖配置
- MySQL 和 Redis 基础设施初始化
- JWT 管理器和鉴权中间件
- Request ID、Recovery、访问日志和 CORS 中间件
- 统一 JSON 响应、错误映射、分页和请求校验工具
- 健康检查和就绪检查接口
- Swagger 生成和 `/swagger/*any`
- `migrations/` 下的数据库迁移
- Makefile 和 GitHub Actions 检查
- 面向 AI 开发工具的 `docs/AI_DEVELOPMENT.md`

## 初始化新项目

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

然后刷新生成文件：

```bash
go mod tidy
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/demo-api/main.go -o docs
go test -buildvcs=false ./...
```

## 运行

```bash
go run ./cmd/warspark server --config=configs/config.dev.yaml
```

默认接口：

- `GET /health`
- `GET /api/v1/health`
- `GET /api/v1/ready`
- `GET /api/v1`
- `GET /swagger/index.html`

## 配置

配置优先级：

```text
defaults -> YAML file -> WARSPARK_* environment variables -> CLI flags
```

示例：

- `WARSPARK_JWT_SECRET`
- `WARSPARK_MYSQL_HOST`
- `WARSPARK_SERVER_PORT`

生产环境密钥应通过环境变量或密钥管理系统注入，不要提交到 YAML 文件中。
数据库名和其他部署相关值通过 YAML 或环境变量配置；初始化脚本只会为了方便更新示例默认值。

## 常用命令

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

## 目录结构

```text
cmd -> app -> api -> controller -> service -> repository -> model
```

命令解析放在 `cmd`，启动和依赖组装放在 `internal/app`，HTTP 绑定放在 `internal/controller`，业务逻辑放在 `internal/service`，持久化访问放在 `internal/repository`，表结构相关模型放在 `internal/model`。

更多说明见 [docs/design.md](docs/design.md)、[docs/template-spec.md](docs/template-spec.md) 和 [docs/AI_DEVELOPMENT.md](docs/AI_DEVELOPMENT.md)。
