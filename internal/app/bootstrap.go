package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"

	v1 "github.com/ww1489/WarSpark/internal/api/v1"
	appconfig "github.com/ww1489/WarSpark/internal/config"
	applogger "github.com/ww1489/WarSpark/internal/infra/logger"
	inframysql "github.com/ww1489/WarSpark/internal/infra/mysql"
	infraredis "github.com/ww1489/WarSpark/internal/infra/redis"
	corsmw "github.com/ww1489/WarSpark/internal/middleware/cors"
	loggingmw "github.com/ww1489/WarSpark/internal/middleware/logging"
	recoverymw "github.com/ww1489/WarSpark/internal/middleware/recovery"
	requestidmw "github.com/ww1489/WarSpark/internal/middleware/requestid"
	"github.com/ww1489/WarSpark/internal/repository"
	"github.com/ww1489/WarSpark/internal/service"
	"github.com/ww1489/WarSpark/internal/worker"
	appjwt "github.com/ww1489/WarSpark/pkg/jwt"
)

type StartOptions struct {
	ConfigPath   string
	PortOverride int
}

// Start initializes infrastructure, starts the HTTP server, and shuts down gracefully.
func Start(parent context.Context, opts StartOptions) error {
	appCtx, cancelApp := context.WithCancel(parent)
	defer cancelApp()

	cfg, err := appconfig.Load(opts.ConfigPath, opts.PortOverride)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, closeLogger, err := applogger.New(cfg.Log)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer closeLogger()

	gin.SetMode(cfg.Server.Mode)

	infraCtx, cancelInfra := context.WithTimeout(appCtx, cfg.Server.StartupTimeout)
	defer cancelInfra()

	mysqlDB, err := inframysql.New(infraCtx, cfg.MySQL)
	if err != nil {
		return fmt.Errorf("init mysql: %w", err)
	}
	defer func() {
		if err := mysqlDB.Close(); err != nil {
			logger.Warn("mysql close failed", slog.String("error", err.Error()))
		}
	}()

	redisClient, err := infraredis.New(infraCtx, cfg.Redis)
	if err != nil {
		return fmt.Errorf("init redis: %w", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Warn("redis close failed", slog.String("error", err.Error()))
		}
	}()

	tokenManager := appjwt.New(appjwt.Config{
		Secret:        cfg.JWT.Secret,
		Issuer:        cfg.JWT.Issuer,
		AccessExpire:  cfg.JWT.AccessExpire,
		RefreshExpire: cfg.JWT.RefreshExpire,
	})

	router := gin.New()
	router.Use(
		requestidmw.New(),
		recoverymw.New(logger),
		loggingmw.AccessLogger(logger),
		corsmw.New(cfg.CORS),
	)

	runtimeConfig := appconfig.NewRuntimeConfig(cfg, logger, mysqlDB, redisClient, tokenManager)
	v1.SetupRoutes(router, runtimeConfig)

	imageSearchRepository := repository.NewImageSearchRepository(mysqlDB)
	imageStorage := service.NewLocalImageStorage("data/uploads", "/uploads")
	imageSearchService := service.NewImageSearchService(imageSearchRepository, imageStorage)
	imageSearchWorker := worker.NewImageSearchWorker(imageSearchService, logger, worker.ImageSearchWorkerOptions{})
	imageSearchWorker.Start(appCtx)

	return runHTTPServer(appCtx, logger, cfg.Server, router)
}
