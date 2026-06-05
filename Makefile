.PHONY: build check docs fmt migrate-create migrate-down migrate-up migrate-version run test tidy vet

# Override these variables when reusing the template, for example:
# make run APP=demo-api CONFIG=configs/config.prod.yaml
APP ?= warspark
CONFIG ?= configs/config.dev.yaml
MIGRATIONS ?= migrations
MIGRATION_NAME ?= change_me
STEPS ?= 1

# Build the service binary from cmd/$(APP).
build:
	go build -buildvcs=false ./cmd/$(APP)

# Run the standard local quality gate.
check: fmt vet test build

# Regenerate Swagger files under docs/.
docs:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/$(APP)/main.go -o docs

# Format all Go packages in place.
fmt:
	gofmt -w .

# Create a new sequential SQL migration pair.
migrate-create:
	go run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1 create -ext sql -dir $(MIGRATIONS) -seq $(MIGRATION_NAME)

# Roll back database migrations. Set STEPS=N to revert more than one migration.
migrate-down:
	go run ./cmd/$(APP) --config=$(CONFIG) migrate down --steps=$(STEPS) --path=$(MIGRATIONS)

# Apply pending database migrations.
migrate-up:
	go run ./cmd/$(APP) --config=$(CONFIG) migrate up --path=$(MIGRATIONS)

# Print the current database migration version.
migrate-version:
	go run ./cmd/$(APP) --config=$(CONFIG) migrate version --path=$(MIGRATIONS)

# Start the HTTP server with the selected config file.
run:
	go run ./cmd/$(APP) server --config=$(CONFIG)

# Run all Go tests with VCS stamping disabled for template-friendly builds.
test:
	go test -buildvcs=false ./...

# Synchronize go.mod and go.sum.
tidy:
	go mod tidy

# Run Go static analysis.
vet:
	go vet ./...
