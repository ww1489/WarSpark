package logging

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/middleware/requestid"
)

func AccessLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		attrs := []slog.Attr{
			slog.String("request_id", requestID(c)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Int64("latency_ms", time.Since(startedAt).Milliseconds()),
			slog.String("client_ip", c.ClientIP()),
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("error", c.Errors.Last().Error()))
		}

		logger.LogAttrs(c.Request.Context(), levelForStatus(c.Writer.Status()), "request", attrs...)
	}
}

func levelForStatus(status int) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

func requestID(c *gin.Context) string {
	value, ok := c.Get(requestid.Key)
	if !ok {
		return ""
	}
	requestID, _ := value.(string)
	return requestID
}
