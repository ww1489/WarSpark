package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	appconfig "github.com/ww1489/WarSpark/internal/config"
)

func New(cfg appconfig.LogConfig) (*slog.Logger, func(), error) {
	level, err := parseLogLevel(cfg.Level)
	if err != nil {
		return nil, nil, err
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				if t, ok := attr.Value.Any().(time.Time); ok {
					attr.Value = slog.StringValue(t.Format("2006-01-02 15:04:05"))
				}
			}
			return attr
		},
	}

	handlers := make([]slog.Handler, 0, 2)
	closers := make([]io.Closer, 0, 1)

	if cfg.Console.Enabled {
		handlers = append(handlers, slog.NewTextHandler(os.Stdout, opts))
	}

	if cfg.File.Enabled {
		if err := os.MkdirAll(filepath.Dir(cfg.File.Path), 0o755); err != nil {
			return nil, nil, fmt.Errorf("create log dir: %w", err)
		}
		writer := &lumberjack.Logger{
			Filename:   cfg.File.Path,
			MaxSize:    cfg.File.MaxSizeMB,
			MaxBackups: cfg.File.MaxBackups,
			MaxAge:     cfg.File.MaxAgeDays,
			Compress:   cfg.File.Compress,
		}
		closers = append(closers, writer)
		handlers = append(handlers, slog.NewJSONHandler(writer, opts))
	}

	logger := slog.New(newMultiHandler(handlers...))
	slog.SetDefault(logger)

	closeFn := func() {
		for _, closer := range closers {
			_ = closer.Close()
		}
	}
	return logger, closeFn, nil
}

func parseLogLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unsupported log level %q", value)
	}
}
