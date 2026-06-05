package recovery

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/middleware/requestid"
	"github.com/ww1489/WarSpark/internal/utils"
)

func New(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error(
					"panic recovered",
					slog.String("request_id", requestID(c)),
					slog.Any("panic", recovered),
				)
				utils.JSON(c, http.StatusInternalServerError, 10006, "internal server error", nil)
				c.Abort()
			}
		}()
		c.Next()
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
