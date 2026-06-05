//go:build !windows

package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/fvbock/endless"

	appconfig "github.com/ww1489/WarSpark/internal/config"
)

func runHTTPServer(ctx context.Context, logger *slog.Logger, cfg appconfig.ServerConfig, handler http.Handler) error {
	addr := fmt.Sprintf(":%d", cfg.Port)
	endless.DefaultHammerTime = cfg.ShutdownTimeout

	server := endless.NewServer(addr, handler)
	server.ReadHeaderTimeout = cfg.ReadHeaderTimeout
	server.ReadTimeout = cfg.ReadTimeout
	server.WriteTimeout = cfg.WriteTimeout
	server.IdleTimeout = cfg.IdleTimeout
	server.BeforeBegin = func(addr string) {
		logger.Info(
			"http server starting",
			slog.String("server", "endless"),
			slog.Int("pid", os.Getpid()),
			slog.String("addr", addr),
			slog.String("mode", cfg.Mode),
		)
	}

	registerEndlessHook(logger, server, endless.PRE_SIGNAL, syscall.SIGHUP, "graceful restart signal received")
	registerEndlessHook(logger, server, endless.PRE_SIGNAL, syscall.SIGINT, "shutdown signal received")
	registerEndlessHook(logger, server, endless.PRE_SIGNAL, syscall.SIGTERM, "shutdown signal received")
	registerEndlessHook(logger, server, endless.POST_SIGNAL, syscall.SIGINT, "shutdown signal handled")
	registerEndlessHook(logger, server, endless.POST_SIGNAL, syscall.SIGTERM, "shutdown signal handled")

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("context canceled; signaling endless server shutdown")
		if process, err := os.FindProcess(os.Getpid()); err == nil {
			_ = process.Signal(syscall.SIGTERM)
		}
		err := <-errCh
		if isExpectedEndlessClose(err) {
			logger.Info("http server stopped")
			return nil
		}
		return fmt.Errorf("http server: %w", err)
	case err := <-errCh:
		if isExpectedEndlessClose(err) {
			logger.Info("http server stopped")
			return nil
		}
		return fmt.Errorf("http server: %w", err)
	}
}

type signalHookRegistrar interface {
	RegisterSignalHook(int, os.Signal, func()) error
}

func registerEndlessHook(logger *slog.Logger, server signalHookRegistrar, phase int, sig os.Signal, message string) {
	if err := server.RegisterSignalHook(phase, sig, func() {
		logger.Info(message, slog.String("signal", sig.String()), slog.Int("pid", os.Getpid()))
	}); err != nil {
		logger.Warn("register endless signal hook failed", slog.String("signal", sig.String()), slog.String("error", err.Error()))
	}
}

func isExpectedEndlessClose(err error) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, net.ErrClosed) {
		return true
	}
	return strings.Contains(err.Error(), "use of closed network connection")
}
