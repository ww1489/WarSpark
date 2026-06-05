package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultsAreValid(t *testing.T) {
	cfg := Defaults()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("defaults should be valid: %v", err)
	}
}

func TestDefaultConfigPathCanBeMissing(t *testing.T) {
	cfg, err := Load("configs/config.yaml", 0)
	if err != nil {
		t.Fatalf("default config path should fall back to defaults: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Fatalf("unexpected default port: %d", cfg.Server.Port)
	}
}

func TestLoadReadsFileAndEnvOverrides(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte(`
server:
  port: 9090
  mode: test
mysql:
  user: file-user
  host: db.internal
  port: "3307"
  db: file_db
jwt:
  secret: file-secret
  issuer: file-issuer
  access_expire: 10m
  refresh_expire: 24h
log:
  console:
    enabled: true
  file:
    enabled: false
`)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	t.Setenv("WARSPARK_SERVER_PORT", "7070")
	t.Setenv("WARSPARK_CORS_ALLOWED_ORIGINS", "https://a.example,https://b.example")
	t.Setenv("WARSPARK_MYSQL_DB", "env_db")
	t.Setenv("WARSPARK_JWT_ACCESS_EXPIRE", "45m")

	cfg, err := Load(configPath, 0)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Port != 7070 {
		t.Fatalf("expected env port override, got %d", cfg.Server.Port)
	}
	if cfg.Server.Mode != "test" {
		t.Fatalf("expected file mode, got %s", cfg.Server.Mode)
	}
	if cfg.MySQL.User != "file-user" || cfg.MySQL.Host != "db.internal" {
		t.Fatalf("expected mysql config from file, got user=%s host=%s", cfg.MySQL.User, cfg.MySQL.Host)
	}
	if cfg.MySQL.DB != "env_db" {
		t.Fatalf("expected mysql db from env, got %s", cfg.MySQL.DB)
	}
	if cfg.JWT.AccessExpire != 45*time.Minute {
		t.Fatalf("expected env duration override, got %s", cfg.JWT.AccessExpire)
	}
	if len(cfg.CORS.AllowedOrigins) != 2 || cfg.CORS.AllowedOrigins[0] != "https://a.example" || cfg.CORS.AllowedOrigins[1] != "https://b.example" {
		t.Fatalf("expected env list override, got %#v", cfg.CORS.AllowedOrigins)
	}
}

func TestMissingCustomConfigPathReturnsError(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.yaml"), 0)
	if err == nil {
		t.Fatal("expected missing custom config path to return error")
	}
}

func TestLoadAppliesPortOverride(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte(`
server:
  port: 9090
  mode: test
mysql:
  user: file-user
  host: db.internal
  port: "3307"
  db: file_db
jwt:
  secret: file-secret
  issuer: file-issuer
log:
  console:
    enabled: true
  file:
    enabled: false
`)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	t.Setenv("WARSPARK_SERVER_PORT", "7070")

	cfg, err := Load(configPath, 6060)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Server.Port != 6060 {
		t.Fatalf("expected port override, got %d", cfg.Server.Port)
	}
}

func TestLoadRejectsInvalidPortOverride(t *testing.T) {
	_, err := Load("configs/config.yaml", 70000)
	if err == nil {
		t.Fatal("expected invalid port override to return error")
	}
}
