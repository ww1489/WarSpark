package mysql

import (
	"strings"
	"testing"

	appconfig "github.com/ww1489/WarSpark/internal/config"
)

func TestMigrationURLUsesMySQLScheme(t *testing.T) {
	cfg := appconfig.MySQLConfig{
		User:      "user",
		Password:  "pass",
		Host:      "127.0.0.1",
		Port:      "3306",
		DB:        "warspark",
		Charset:   "utf8mb4",
		ParseTime: true,
		Loc:       "Local",
	}

	url := MigrationURL(cfg)
	if !strings.HasPrefix(url, "mysql://") {
		t.Fatalf("expected mysql scheme, got %s", url)
	}
	if !strings.Contains(url, "warspark") {
		t.Fatalf("expected database name in url, got %s", url)
	}
}
