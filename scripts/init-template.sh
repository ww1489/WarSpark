#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  scripts/init-template.sh \
    --module-path github.com/acme/demo-api \
    --slug demo-api \
    --display-name "Demo API" \
    [--env-prefix DEMO_API]
EOF
}

module_path=""
slug=""
display_name=""
env_prefix=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --module-path)
      module_path="${2:-}"
      shift 2
      ;;
    --slug)
      slug="${2:-}"
      shift 2
      ;;
    --display-name)
      display_name="${2:-}"
      shift 2
      ;;
    --env-prefix)
      env_prefix="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ -z "$module_path" || -z "$slug" || -z "$display_name" ]]; then
  usage >&2
  exit 1
fi

if [[ ! "$slug" =~ ^[a-z][a-z0-9-]*$ ]]; then
  echo "Slug must start with a lowercase letter and contain only lowercase letters, numbers, and hyphens." >&2
  exit 1
fi

if [[ -z "$env_prefix" ]]; then
  env_prefix="$(printf '%s' "$slug" | tr '[:lower:]-' '[:upper:]_')"
fi
if [[ ! "$env_prefix" =~ ^[A-Z][A-Z0-9_]*$ ]]; then
  echo "EnvPrefix must start with an uppercase letter and contain only uppercase letters, numbers, and underscores." >&2
  exit 1
fi

# Keep generated MySQL sample identifiers SQL-friendly when a service slug uses hyphens.
# This is not an init parameter; runtime config still owns mysql.user and mysql.db.
database_name="${slug//-/_}"

current_module_path="github.com/ww1489/WarSpark"
current_slug="warspark"
current_display_name="WarSpark"
current_env_prefix="WARSPARK"

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"
cd "$repo_root"

old_command_path="cmd/$current_slug"
new_command_path="cmd/$slug"
if [[ -d "$old_command_path" && "$slug" != "$current_slug" ]]; then
  if [[ -e "$new_command_path" ]]; then
    echo "Target command path already exists: $new_command_path" >&2
    exit 1
  fi
  mv "$old_command_path" "$new_command_path"
fi

replace_file() {
  local file="$1"
  local rel="$2"
  CURRENT_MODULE_PATH="$current_module_path" \
  MODULE_PATH="$module_path" \
  CURRENT_DISPLAY_NAME="$current_display_name" \
  DISPLAY_NAME="$display_name" \
  CURRENT_ENV_PREFIX="$current_env_prefix" \
  ENV_PREFIX="$env_prefix" \
  CURRENT_SLUG="$current_slug" \
  SLUG="$slug" \
    perl -0pi \
      -e 'BEGIN { $module = quotemeta($ENV{"CURRENT_MODULE_PATH"}); $display = quotemeta($ENV{"CURRENT_DISPLAY_NAME"}); $env_prefix = quotemeta($ENV{"CURRENT_ENV_PREFIX"}); $slug = quotemeta($ENV{"CURRENT_SLUG"}); } s/$module/$ENV{"MODULE_PATH"}/g; s/$display/$ENV{"DISPLAY_NAME"}/g; s/$env_prefix/$ENV{"ENV_PREFIX"}/g; s/$slug/$ENV{"SLUG"}/g;' \
      "$file"

  # Normalize only sample/config defaults. Deployments override these through YAML or env vars.
  if [[ "$database_name" != "$slug" && "$rel" != "scripts/init-template.ps1" && "$rel" != "scripts/init-template.sh" ]]; then
    SLUG="$slug" \
    DATABASE_NAME="$database_name" \
      perl -0pi \
        -e 'BEGIN { $slug = quotemeta($ENV{"SLUG"}); $database = $ENV{"DATABASE_NAME"}; $db_value = quotemeta("DB:              \"" . $ENV{"SLUG"} . "\""); } s/user: $slug/user: $database/g; s/db: $slug/db: $database/g; s/DB:              AppName/DB:              "$database"/g; s/$db_value/DB:              "$database"/g;' \
        "$file"
  fi
}

while IFS= read -r -d '' file; do
  rel="${file#./}"
  case "$rel" in
    .git/*|go.sum|docs/swagger.json|docs/swagger.yaml)
      continue
      ;;
  esac

  case "$rel" in
    Makefile|LICENSE|*.go|*.mod|*.md|*.yaml|*.yml|*.ps1|*.sh|*.gitignore)
      replace_file "$file" "$rel"
      ;;
  esac
done < <(find . -type f -print0)

cat <<EOF
Template initialized:
  module: $module_path
  command: $slug
  display: $display_name
  env prefix: $env_prefix
  sample mysql identifiers: $database_name (override in config)

Next steps:
  go mod tidy
  go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/$slug/main.go -o docs
  go test -buildvcs=false ./...
EOF
