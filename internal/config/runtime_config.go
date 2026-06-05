package config

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	appjwt "github.com/ww1489/WarSpark/pkg/jwt"
)

type RuntimeConfig struct {
	Config       Config
	Logger       *slog.Logger
	MySQL        *sqlx.DB
	Redis        *redis.Client
	TokenManager *appjwt.Manager
}

func NewRuntimeConfig(cfg Config, logger *slog.Logger, mysqlDB *sqlx.DB, redisClient *redis.Client, tokenManager *appjwt.Manager) RuntimeConfig {
	return RuntimeConfig{
		Config:       cfg,
		Logger:       logger,
		MySQL:        mysqlDB,
		Redis:        redisClient,
		TokenManager: tokenManager,
	}
}
