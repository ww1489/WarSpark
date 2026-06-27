package config_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	goredis "github.com/redis/go-redis/v9"

	_ "github.com/go-sql-driver/mysql"

	appconfig "github.com/ww1489/WarSpark/internal/config"
)

func TestMySQLConnection(t *testing.T) {
	cfg, err := appconfig.Load("../../configs/config.dev.yaml", 0)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	dsn := mysqlDSN(cfg.MySQL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sqlx.ConnectContext(ctx, "mysql", dsn)
	if err != nil {
		t.Fatalf("connect mysql: %v\nDSN: %s", err, dsn)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping mysql: %v", err)
	}
	t.Logf("MySQL connected: %s:%s/%s", cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.DB)
}

func TestRedisConnection(t *testing.T) {
	cfg, err := appconfig.Load("../../configs/config.dev.yaml", 0)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	client := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Redis.Addr + ":" + cfg.Redis.Port,
		Username: cfg.Redis.User,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		t.Fatalf("connect redis: %v\nAddr: %s:%s", err, cfg.Redis.Addr, cfg.Redis.Port)
	}
	defer client.Close()
	t.Logf("Redis connected: %s:%s", cfg.Redis.Addr, cfg.Redis.Port)
}

func mysqlDSN(cfg appconfig.MySQLConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB, cfg.Charset, cfg.ParseTime,
	)
}
