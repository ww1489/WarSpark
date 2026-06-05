package requestid

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const Key = "request_id"

func New() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set(Key, requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}
