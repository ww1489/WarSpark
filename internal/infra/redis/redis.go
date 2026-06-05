package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	appconfig "github.com/ww1489/WarSpark/internal/config"
)

func New(ctx context.Context, cfg appconfig.RedisConfig) (*goredis.Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:            fmt.Sprintf("%s:%s", cfg.Addr, cfg.Port),
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}
